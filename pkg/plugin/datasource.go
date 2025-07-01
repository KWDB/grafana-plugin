package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kwdb/kwdb/pkg/models"
)

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler interfaces. Plugin should not implement all these
// interfaces - only those which are required for a particular task.
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

// NewDatasource creates a new datasource instance.
func NewDatasource(_ context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	config, err := models.LoadPluginSettings(settings)
	if err != nil {
		return nil, fmt.Errorf("Failed to load settings: %v", err)
	}

	url := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		config.Username, config.Secrets.Password, config.Host, config.Port, config.Database)

	poolConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse connection configuration: %v", err)
	}

	poolConfig.MaxConns = 20                               // 最大连接数（避免数据库过载）
	poolConfig.MinConns = 2                                // 最小空闲连接数（加速突发请求）
	poolConfig.MaxConnLifetime = 10 * time.Minute          // 连接最大存活时间（防内存泄漏）
	poolConfig.MaxConnIdleTime = 5 * time.Minute           // 空闲连接超时时间（释放资源）
	poolConfig.HealthCheckPeriod = 1 * time.Minute         // 健康检查间隔（检测失效连接）
	poolConfig.ConnConfig.ConnectTimeout = 5 * time.Second // 连接建立超时[2,4](@ref)

	poolConfig.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
		log.DefaultLogger.Info("Acquiring connection from pool")
		return true
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to create connection pool: %v", err)
	}

	return &Datasource{
		config: config,
		pool:   pool,
	}, nil
}

type Datasource struct {
	config *models.PluginSettings
	pool   *pgxpool.Pool
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *Datasource) Dispose() {
	// Clean up datasource instance resources.
	d.pool.Close()
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()
	response.Responses = make(map[string]backend.DataResponse, len(req.Queries))

	var (
		wg      sync.WaitGroup
		resLock sync.Mutex
		errOnce sync.Once
		error   error
	)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, q := range req.Queries {
		wg.Add(1)
		go func(query backend.DataQuery) {
			defer wg.Done()

			queryCtx, timeoutCancel := context.WithTimeout(ctx, query.TimeRange.Duration())
			defer timeoutCancel()

			log.DefaultLogger.Info("Processing query", "refID", query.RefID)

			// 获取连接并自动释放
			conn, err := d.pool.Acquire(queryCtx)
			if err != nil {
				errOnce.Do(func() {
					error = fmt.Errorf("connection acquisition failed for %s: %w", query.RefID, err)
					cancel()
				})
				return
			}
			defer conn.Release()

			// 执行查询逻辑
			result := d.query(queryCtx, req.PluginContext, query)

			// 线程安全写入响应
			resLock.Lock()
			response.Responses[query.RefID] = result
			resLock.Unlock()
		}(q)
	}

	// 异步等待所有goroutine完成
	go func() {
		wg.Wait()
		cancel()
	}()

	// 监听上下文取消信号
	<-ctx.Done()

	if ctx.Err() == context.Canceled && error != nil {
		return nil, error
	}

	return response, nil
}

type queryModel struct {
	QueryText string `json:"queryText"`
}

func (d *Datasource) query(ctx context.Context, _ backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	var response backend.DataResponse
	qm, err := getQueryModel(query)
	if err != nil {
		log.DefaultLogger.Error("Failed to parse query model", "error", err)
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("invalid query format: %v", err))
	}

	if qm.QueryText == "" {
		return backend.ErrDataResponse(backend.StatusBadRequest, "empty query text")
	}

	result, err := queryDataFromDatasource(ctx, d.pool, qm.QueryText)
	if err != nil {
		log.DefaultLogger.Error("Query execution failed", "query", qm.QueryText, "error", err)
		return backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("query execution failed: %v", err))
	}

	frame := formatData(result, qm.QueryText, query.RefID)

	response.Frames = append(response.Frames, frame)
	return response
}

// getQueryModel parses the query JSON and performs variable substitution
func getQueryModel(query backend.DataQuery) (*queryModel, error) {
	var qm queryModel
	if err := json.Unmarshal(query.JSON, &qm); err != nil {
		return nil, err
	}

	// Perform variable substitution
	qm.QueryText = strings.ReplaceAll(qm.QueryText, "$from", query.TimeRange.From.Format(time.RFC3339))
	qm.QueryText = strings.ReplaceAll(qm.QueryText, "$to", query.TimeRange.To.Format(time.RFC3339))
	qm.QueryText = strings.ReplaceAll(qm.QueryText, "$interval", query.Interval.String())

	return &qm, nil
}

// queryDataFromDatasource executes the query and returns the raw results
func queryDataFromDatasource(ctx context.Context, pool *pgxpool.Pool, queryText string) ([]map[string]interface{}, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	rows, err := conn.Query(ctx, queryText)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fieldDescs := rows.FieldDescriptions()
	var results []map[string]interface{}

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, fd := range fieldDescs {
			row[string(fd.Name)] = values[i]
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// formatData formats standard query results into data.Frame
func formatData(results []map[string]interface{}, queryText, refID string) *data.Frame {
	if len(results) == 0 {
		return data.NewFrame("response")
	}

	// 收集所有字段的类型（按列处理）
	fieldTypes := make(map[string]data.FieldType)
	for _, row := range results {
		for k, v := range row {
			if _, ok := fieldTypes[k]; !ok {
				fieldTypes[k] = data.FieldTypeFor(v)
			}
		}
	}

	// 创建字段并统一类型（优先使用出现次数最多的类型）
	fields := make([]*data.Field, 0, len(fieldTypes))
	for k, t := range fieldTypes {
		field := data.NewFieldFromFieldType(t, 0)
		field.Name = k
		fields = append(fields, field)
	}

	// 填充数据时进行类型转换
	for _, row := range results {
		for _, field := range fields {
			v, ok := row[field.Name]
			if !ok {
				continue
			}
			// 类型转换逻辑（示例：字符串转float64）
			if field.Type() == data.FieldTypeFloat64 && reflect.TypeOf(v).Kind() == reflect.String {
				floatVal, err := strconv.ParseFloat(v.(string), 64)
				if err == nil {
					v = floatVal
				}
			}
			field.Append(v)
		}
	}

	frame := data.NewFrame("response", fields...)

	frame.Meta = &data.FrameMeta{
		ExecutedQueryString: queryText,
		Type:                data.FrameTypeTimeSeriesWide,
		Custom: map[string]interface{}{
			"data": map[string]interface{}{
				"requestId": refID,
			},
		},
	}

	return frame
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	// Verify settings first
	config, err := models.LoadPluginSettings(*req.PluginContext.DataSourceInstanceSettings)
	if err != nil {
		log.DefaultLogger.Error("Failed to load plugin settings", "error", err)
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Invalid datasource configuration",
		}, nil
	}

	// Check connection status
	if d.pool == nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Connection not initialized",
		}, nil
	}

	// Execute simple query with timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := d.pool.Acquire(ctx)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: handleConnectionErrorMessage(err, d.pool.Config().ConnConfig),
		}, nil
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SELECT 1"); err != nil {
		log.DefaultLogger.Error("Database health check failed",
			"host", config.Host, "port", config.Port, "error", err)
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Database query failed: %v", err),
		}, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source is working",
	}, nil
}

func handleConnectionErrorMessage(err error, cfg *pgx.ConnConfig) string {
	if strings.Contains(err.Error(), "password authentication failed") {
		return "Password authentication failed: please check your username/password"
	}
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return "Connection timed out: Please check the host address and port"
	}
	if strings.Contains(err.Error(), "connection refused") {
		if cfg.Port < 1 || cfg.Port > 65535 {
			return "Invalid port: port range should be 1-65535"
		}
		return "Connection refused: please check the host port and firewall settings"
	}

	return fmt.Sprintf("KWDB connection failed: %v", err)
}
