package plugin

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kwdb/kwdb/pkg/models"
)

func TestHandleConnectionErrorMessage(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		cfg      *pgx.ConnConfig
		expected string
	}{
		{
			name:     "password authentication failed",
			err:      errors.New("password authentication failed for user"),
			cfg:      &pgx.ConnConfig{Config: pgconn.Config{Host: "localhost", Port: 5432}},
			expected: "Password authentication failed: please check your username/password",
		},
		{
			name:     "connection timeout",
			err:      &net.DNSError{IsTimeout: true},
			cfg:      &pgx.ConnConfig{Config: pgconn.Config{Host: "localhost", Port: 5432}},
			expected: "Connection timed out: Please check the host address and port",
		},
		{
			name:     "connection refused with invalid port",
			err:      errors.New("connection refused"),
			cfg:      &pgx.ConnConfig{Config: pgconn.Config{Host: "localhost", Port: 0}},
			expected: "Invalid port: port range should be 1-65535",
		},
		{
			name:     "connection refused with valid port",
			err:      errors.New("connection refused"),
			cfg:      &pgx.ConnConfig{Config: pgconn.Config{Host: "localhost", Port: 5432}},
			expected: "Connection refused: please check the host port and firewall settings",
		},
		{
			name:     "other errors",
			err:      errors.New("unknown error"),
			cfg:      &pgx.ConnConfig{Config: pgconn.Config{Host: "localhost", Port: 5432}},
			expected: "KWDB connection failed: unknown error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := handleConnectionErrorMessage(tc.err, tc.cfg)
			if result != tc.expected {
				t.Errorf("expected message: %q, got: %q", tc.expected, result)
			}
		})
	}
}

func TestCheckHealth(t *testing.T) {
	t.Run("invalid settings", func(t *testing.T) {
		req := &backend.CheckHealthRequest{
			PluginContext: backend.PluginContext{
				DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
					JSONData: []byte(`invalid json`),
				},
			},
		}
		d := &Datasource{}
		result, err := d.CheckHealth(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Status != backend.HealthStatusError {
			t.Errorf("expected status Error, got %v", result.Status)
		}
		if result.Message != "Invalid datasource configuration" {
			t.Errorf("unexpected message: %q", result.Message)
		}
	})

	t.Run("pool is nil", func(t *testing.T) {
		d := &Datasource{pool: nil}
		req := &backend.CheckHealthRequest{
			PluginContext: backend.PluginContext{
				DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
					JSONData:                []byte(`{"host":"localhost","port":5432,"database":"test","username":"user"}`),
					DecryptedSecureJSONData: map[string]string{"password": "pass"},
				},
			},
		}
		result, err := d.CheckHealth(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Status != backend.HealthStatusError {
			t.Errorf("expected status Error, got %v", result.Status)
		}
		if result.Message != "Connection not initialized" {
			t.Errorf("unexpected message: %q", result.Message)
		}
	})

	t.Run("acquire error", func(t *testing.T) {
		url := "postgresql://user:pass@localhost:5432/test?sslmode=disable"
		cfg, err := pgxpool.ParseConfig(url)
		if err != nil {
			t.Fatal(err)
		}
		cfg.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
			return false
		}
		pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer pool.Close()

		d := &Datasource{
			pool: pool,
			config: &models.PluginSettings{
				Host: "localhost",
				Port: 5432,
			},
		}
		req := &backend.CheckHealthRequest{
			PluginContext: backend.PluginContext{
				DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
					JSONData:                []byte(`{"host":"localhost","port":5432,"database":"test","username":"user"}`),
					DecryptedSecureJSONData: map[string]string{"password": "pass"},
				},
			},
		}
		result, err := d.CheckHealth(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Status != backend.HealthStatusError {
			t.Errorf("expected status Error, got %v", result.Status)
		}
	})
}

func TestGetQueryModel(t *testing.T) {
	testCases := []struct {
		name        string
		queryJSON   string
		from        time.Time
		to          time.Time
		interval    time.Duration
		expectedSQL string
		expectError bool
	}{{
		name: "normal variable substitution",
		queryJSON: `{"queryText": "SELECT * FROM metrics WHERE time > '$from' AND time < '$to' AND interval = '$interval'"}`,
		from: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		to: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
		interval: time.Minute * 5,
		expectedSQL: "SELECT * FROM metrics WHERE time > '2023-01-01T00:00:00Z' AND time < '2023-01-02T00:00:00Z' AND interval = '5m0s'",
		expectError: false,
	}, {
		name: "invalid JSON format",
		queryJSON: `{invalid json}`,
		expectError: true,
	}, {
		name: "no variables to substitute",
		queryJSON: `{"queryText": "SELECT * FROM metrics"}`,
		from: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		to: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
		interval: time.Minute * 5,
		expectedSQL: "SELECT * FROM metrics",
		expectError: false,
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query := backend.DataQuery{
				JSON: []byte(tc.queryJSON),
				TimeRange: backend.TimeRange{
					From: tc.from,
					To: tc.to,
				},
				Interval: tc.interval,
			}

			model, err := getQueryModel(query)

			if tc.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if model.QueryText != tc.expectedSQL {
				t.Errorf("expected SQL: %q, got: %q", tc.expectedSQL, model.QueryText)
			}
		})
	}
}

func TestFormatData(t *testing.T) {
	testCases := []struct {
		name     string
		results  []map[string]interface{}
		query    string
		refID    string
		validate func(*testing.T, *data.Frame)
	}{
		{
			name: "basic conversion",
			results: []map[string]interface{}{
				{
					"current":   5.1,
					"meter_id":  "M2",
					"power":     int64(1050),
					"rule_name": "高压告警",
					"ts":        "2025-04-16T07:40:22.284Z",
					"voltage":   221.0,
				},
				{
					"current":   5.1,
					"meter_id":  "M2",
					"power":     int64(1050),
					"rule_name": "低压告警",
					"ts":        "2025-04-16T07:40:22.284Z",
					"voltage":   221.0,
				},
			},
			query: "test query",
			refID: "A",
			validate: func(t *testing.T, frame *data.Frame) {
				if frame.Rows() != 2 {
					t.Errorf("expected 2 rows, got %d", frame.Rows())
				}

				expectedFields := map[string]data.FieldType{
					"current":   data.FieldTypeFloat64,
					"meter_id":  data.FieldTypeString,
					"power":     data.FieldTypeInt64,
					"rule_name": data.FieldTypeString,
					"ts":        data.FieldTypeString,
					"voltage":   data.FieldTypeFloat64,
				}

				for _, field := range frame.Fields {
					expectedType, ok := expectedFields[field.Name]
					if !ok {
						t.Errorf("unexpected field: %s", field.Name)
						continue
					}
					if field.Type() != expectedType {
						t.Errorf("field %s type mismatch: expected %v got %v",
							field.Name, expectedType, field.Type())
					}
				}

				if meta := frame.Meta; meta != nil {
					if meta.ExecutedQueryString != "test query" {
						t.Errorf("unexpected query in meta: %s", meta.ExecutedQueryString)
					}
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			frame := formatData(tc.results, tc.query, tc.refID)
			tc.validate(t, frame)
		})
	}
}
