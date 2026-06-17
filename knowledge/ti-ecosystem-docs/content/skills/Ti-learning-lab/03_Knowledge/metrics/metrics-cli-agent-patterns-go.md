# Nghiên Cứu Metrics & CLI Agent Patterns trong Go

> **Version**: 1.0.0  
> **Last Updated**: 2026-04-28  
> **Language**: Tiếng Việt  
> **Category**: Research  
> **Purpose**: Nghiên cứu metrics và CLI patterns trong Go để implement metrics-agent và cli-agent cho Ti router

---

## Tổng Quan

Nghiên cứu này tập trung vào việc tìm hiểu metrics và CLI patterns trong Go để implement:
- **metrics-agent**: Usage tracking, logging, telemetry
- **cli-agent**: CLI wrappers, PATH, config management

---

## 1. Metrics Agent Patterns

### 1.1 Internal Research - Ti Router Implementation

**Location**: `Z:\Ti\router\layers\`

#### Existing Metrics Files:
- `layers/monitoring/metrics.go` - Metrics collection
- `layers/telemetry/telemetry_streaming.go` - Telemetry streaming
- `agent-store/router-agent/metrics/metrics.go` - Router agent metrics

#### Key Features:
- Request counting
- Latency tracking
- Error tracking
- Token usage tracking
- Provider health metrics

### 1.2 External Research - Go Metrics Libraries

**Popular Libraries**:

| Library | Stars | Features |
|---------|-------|----------|
| **prometheus/client_golang** | 5.2k | Prometheus client, metrics export |
| **influxdata/telegraf** | 14k | Metrics collection agent |
| **beorn7/perks** | 4.5k | Histogram quantiles |
| **go-kit/metrics** | 5.2k | Metrics abstraction layer |

**Recommendation**: Sử dụng **prometheus/client_golang** cho standard metrics export

### 1.3 Best Practices cho Metrics

#### Metrics Types:
```go
// Counter - Monotonically increasing
var requestCount = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "router_requests_total",
        Help: "Total number of requests",
    },
    []string{"method", "endpoint", "status"},
)

// Histogram - Latency distribution
var requestDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "router_request_duration_seconds",
        Help:    "Request duration in seconds",
        Buckets: prometheus.DefBuckets,
    },
    []string{"endpoint"},
)

// Gauge - Current value
var activeConnections = prometheus.NewGauge(
    prometheus.GaugeOpts{
        Name: "router_active_connections",
        Help: "Number of active connections",
    },
)

// Summary - Latency with quantiles
var requestLatency = prometheus.NewSummaryVec(
    prometheus.SummaryOpts{
        Name: "router_request_latency_seconds",
        Help: "Request latency in seconds",
    },
    []string{"endpoint"},
)
```

#### Metrics Collection:
```go
type MetricsCollector struct {
    requests      *prometheus.CounterVec
    duration      *prometheus.HistogramVec
    errors        *prometheus.CounterVec
    tokenUsage    *prometheus.CounterVec
}

func (m *MetricsCollector) RecordRequest(method, endpoint string, status int, duration time.Duration) {
    m.requests.WithLabelValues(method, endpoint, fmt.Sprintf("%d", status)).Inc()
    m.duration.WithLabelValues(endpoint).Observe(duration.Seconds())
}

func (m *MetricsCollector) RecordError(endpoint, errorType string) {
    m.errors.WithLabelValues(endpoint, errorType).Inc()
}

func (m *MetricsCollector) RecordTokenUsage(provider string, tokens int) {
    m.tokenUsage.WithLabelValues(provider).Add(float64(tokens))
}
```

#### Prometheus Export:
```go
func StartMetricsServer(addr string) {
    http.Handle("/metrics", promhttp.Handler())
    log.Fatal(http.ListenAndServe(addr, nil))
}
```

---

## 2. CLI Agent Patterns

### 2.1 Internal Research - Ti CLI

**Location**: `Z:\Ti\`

#### Existing CLI Files:
- `cmd/beads/` - BEADS CLI
- `cmd/beadsviz/` - Beadsviz dashboard
- `Z:\02_CORE\_cli\bin\` - CLI binaries

#### Key Features:
- Cobra framework cho CLI
- Configuration management
- PATH management
- Command execution

### 2.2 External Research - Go CLI Libraries

**Popular Libraries**:

| Library | Stars | Features |
|---------|-------|----------|
| **spf13/cobra** | 38k | CLI framework, commands, flags |
| **urfave/cli** | 22k | CLI framework, commands, flags |
| **alecthomas/kong** | 12k | CLI configuration |
| **mitchellh/cli** | 6.5k | CLI framework (deprecated) |

**Recommendation**: Sử dụng **spf13/cobra** (Ti đang dùng)

### 2.3 Best Practices cho CLI

#### Command Structure:
```go
package cmd

import (
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "ti",
    Short: "Ti AI Agent Platform CLI",
    Long:  `Ti is a multi-module Go-based AI agent ecosystem.`,
}

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print version",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Ti v1.0.0")
    },
}

func init() {
    rootCmd.AddCommand(versionCmd)
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        log.Fatal(err)
    }
}
```

#### Configuration Management:
```go
type Config struct {
    RouterAddr string `yaml:"router_addr"`
    LogLevel    string `yaml:"log_level"`
    Debug       bool   `yaml:"debug"`
}

func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    return &config, nil
}
```

#### PATH Management:
```go
func AddToPATH(path string) error {
    currentPath := os.Getenv("PATH")
    newPath := path + string(os.PathListSeparator) + currentPath
    return os.Setenv("PATH", newPath)
}

func GetExecutablePath() (string, error) {
    execPath, err := os.Executable()
    if err != nil {
        return "", err
    }
    return filepath.Abs(execPath), nil
}
```

---

## 3. Recommendations cho Ti Agents

### 3.1 Metrics Agent

**Implementation Priority**:
1. Prometheus metrics export (/metrics endpoint)
2. Request counting and latency tracking
3. Error tracking
4. Token usage tracking
5. Provider health metrics

**Architecture**:
```
metrics-agent/
├── collector/
│   ├── request.go          // Request metrics
│   ├── latency.go          // Latency metrics
│   ├── error.go            // Error metrics
│   └── token.go            // Token usage metrics
├── export/
│   ├── prometheus.go       // Prometheus export
│   └── middleware.go       // Metrics middleware
└── storage/
    └── influxdb.go         // InfluxDB storage (optional)
```

### 3.2 CLI Agent

**Implementation Priority**:
1. CLI wrapper cho existing tools
2. PATH management
3. Configuration management
4. Command execution
5. Help and documentation

**Architecture**:
```
cli-agent/
├── commands/
│   ├── root.go             // Root command
│   ├── config.go           // Config command
│   ├── path.go             // PATH management
│   └── exec.go             // Command execution
├── config/
│   ├── loader.go           // Config loader
│   └── validator.go        // Config validator
└── wrappers/
    ├── claude.go           // Claude wrapper
    ├── devin.go            // Devin wrapper
    └── router.go           // Router wrapper
```

---

## 4. References

### GitHub Repositories:
- **prometheus/client_golang**: https://github.com/prometheus/client_golang
- **spf13/cobra**: https://github.com/spf13/cobra
- **urfave/cli**: https://github.com/urfave/cli

### Ti Router Files:
- **metrics.go**: Z:\Ti\router\layers\monitoring/metrics.go
- **telemetry_streaming.go**: Z:\Ti\router\layers\telemetry/telemetry_streaming.go

---

*Last Updated: 2026-04-28*
*Research completed by: Claude*
*Purpose: Implement metrics-agent và cli-agent cho Ti router ecosystem*
