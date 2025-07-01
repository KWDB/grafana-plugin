# KWDB Grafana Plugin RoadMap

Currently at **v1.0.0 initial version**. This document describes the development roadmap for the KWDB Grafana plugin, including currently completed features, short-term goals, and long-term objectives.

## 🎯 Current Status (v1.0.0)

### ✅ Completed Features

- **Plugin Foundation Architecture**: Standard Grafana plugin structure with React frontend + Go backend
- **Data Source Configuration**: Configuration interface for host, port, database, and authentication information
- **SQL Query Editor**: Code highlighting, formatting, and execution capabilities
- **KWDB Connection**: Connect to KWDB database via PostgreSQL protocol
- **Basic Testing**: E2E tests covering configuration and query functionality
- **CI/CD**: Automated build, test, and release pipeline

## 🚀 Feature Items

### Query Function Enhancement

- **Grafana Time Macro Support**
  - Improve `$__timeFrom()`, `$__timeTo()` macros
  - Add `$__timeGroup()` time grouping macro
  - Support `$__interval` dynamic intervals
  - Support more Grafana macros

### Configuration Optimization

- **Plugin Configuration Enhancement**
  - Detailed connection test feedback
  - SSL/TLS connection options
  - Connection pool configuration options
  - Query timeout settings

## 📋 Technical Improvement Items

### Plugin Quality Enhancement

- [ ] Increase unit test coverage to 80%+
- [ ] Improve TypeScript type definitions
- [ ] Optimize plugin package size and loading performance
- [ ] Add plugin configuration validation

### User Experience Optimization

- [ ] Improve error messages and user guidance
- [ ] Add plugin usage documentation and examples

### Grafana Ecosystem Integration

- [ ] Support Grafana Alerting
- [ ] Compatible with Grafana Cloud
- [ ] Support Grafana variables and annotations
- [ ] Integrate Grafana Explore functionality

## 🔧 Development Priority

### High Priority

1. **Documentation Improvement** - User adoption
1. **Variable Support** - Grafana integration
1. **Grafana Time Macro Enhancement** - Core plugin functionality
1. **Query Performance Optimization** - Critical user experience

### Medium-Low Priority

1. **SSL Connection Support** - Security requirements
1. **Query Builder** - Usability enhancement
1. **Error Handling Improvement** - Plugin stability
1. **Test Coverage Enhancement** - Code quality assurance