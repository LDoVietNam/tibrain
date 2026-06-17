# Monitoring and Metrics Architecture

This document describes the monitoring and metrics architecture for Ti Router and CLI.

## Overview

Ti Router and CLI use Prometheus for metrics collection and Grafana for visualization. This architecture provides real-time monitoring, alerting, and performance analysis.

## Components

### 1. Prometheus

**Purpose**: Metrics collection and storage

**Installation**: `Z:\09_TOOLS\prometheus`

**Configuration**: `Z:\10_WORKPLACE\Ti\prometheus.yml`

**Port**: 9090

**Web UI**: http://localhost:9090

**Scrape Interval**: 15 seconds

### 2. Grafana

**Purpose**: Metrics visualization and dashboarding

**Installation**: `Z:\09_TOOLS\grafana`

**Configuration**: Auto-provisioned via `conf/provisioning/`

**Port**: 3000

**Web UI**: http://localhost:3000

**Credentials**: admin / admin (change on first login)

### 3. Ti Router

**Purpose**: API router and CLI registry

**Metrics Endpoint**: `/metrics/prometheus`

**Port**: 1807

## Router Metrics

### Request Metrics

- `ti_router_requests_total` - Total requests
- `ti_router_requests_success` - Successful requests
- `ti_router_requests_errors` - Failed requests
- `ti_router_request_rate_rps` - Requests per second
- `ti_router_request_rate_rpm` - Requests per minute

### Latency Metrics

- `ti_router_latency_avg_ms` - Average latency in ms
- `ti_router_latency_p50_ms` - P50 latency in ms
- `ti_router_latency_p95_ms` - P95 latency in ms
- `ti_router_latency_p99_ms` - P99 latency in ms

### Token Metrics

- `ti_router_tokens_prompt` - Total prompt tokens
- `ti_router_tokens_completion` - Total completion tokens
- `ti_router_tokens_total` - Total tokens
- `ti_router_cost_total_usd` - Total cost in USD

### Connection Metrics

- `ti_router_connections_active` - Active connections
- `ti_router_connections_peak` - Peak connections

### CLI Metrics

- `ti_router_cli_registrations_total` - Total CLI registrations
- `ti_router_cli_heartbeats_total` - Total CLI heartbeats
- `ti_router_cli_heartbeats_failed` - Failed CLI heartbeats
- `ti_router_cli_active_count` - Active CLI instances
- `ti_router_cli_discoveries_total` - Total CLI discoveries

## CLI Metrics

### Health Endpoint

CLI provides a health endpoint at `ti health` that returns:

```json
{
  "status": "healthy",
  "timestamp": "2026-05-04T00:20:52Z",
  "cli_id": "cli-instance",
  "version": "3.0.0-ti-ecosystem-go1.23",
  "go_version": "go1.25.0",
  "os_arch": "windows/amd64",
  "memory": {
    "alloc_mb": 1,
    "goroutines": 3,
    "num_gc": 0,
    "sys_mb": 6
  },
  "current_mode": "normal",
  "router_url": "http://localhost:1807",
  "nats_url": "nats://localhost:4222"
}
```

## Alerting Rules

Prometheus alerting rules are defined in `prometheus-alerts.yml`:

### High Error Rate
- **Condition**: `rate(ti_router_requests_errors[5m]) > 0.1`
- **Severity**: warning
- **Threshold**: 0.1 errors/sec

### No Active CLIs
- **Condition**: `ti_router_cli_active_count == 0`
- **Severity**: info
- **Duration**: 5 minutes

### High Heartbeat Failure Rate
- **Condition**: `rate(ti_router_cli_grate_heartbeats_failed[5m]) / rate(ti_router_cli_heartbeats_total[5m]) > 0.5`
- **Severity**: warning
- **Threshold**: 50% failure rate

### High Latency
- **Condition**: `ti_router_latency_p95_ms > 1000`
- **Severity**: warning
- **Threshold**: 1000ms

### Router Down
- **Condition**: `up{job="ti-router"} == 0`
- **Severity**: critical
- **Duration**: 1 minute

## Performance Benchmarks

### Router Performance Test Results

- **Requests**: 100
- **Concurrency**: 10
- **Success Rate**: 100%
- **Avg Latency**: 1.54ms
- **P95 Latency**: 2.55ms
- **Requests/sec**: 6291

### Metrics Endpoint Performance

- **Requests**: 50
- **Concurrency**: 5
- **Avg Latency**: 0.79ms
- **P95 Latency**: 1.11ms
- **Requests/sec**: 5954

## Docker Deployment

### Docker Compose

Services:
- **ti-router**: Router service
- **ti-prometheus**: Prometheus metrics collector
- **ti-grafana**: Grafana visualization

Start with:
```bash
docker-compose up -d
```

Stop with:
```bash
docker-compose down
```

See `DOCKER_SETUP.md` for detailed instructions.

## Integration Tests

### CLI Integration Tests

Location: `apps/cli/internal/integration/`

Tests:
- Router helper tests
- Registration tests
- Heartbeat tests
- Discovery tests
- NATS fallback tests
- Performance tests

### Router Integration Tests

Location: `apps/router/layers/metrics/`

Tests:
- CLI metrics tests
- Metrics recording tests

## Troubleshooting

### Prometheus Not Scraping Router

1. Check Router is running:
   ```bash
   curl http://localhost:1807/health
   ```

2. Check metrics endpoint:
   ```bash
   curl http://localhost:1807/metrics/prometheus
   ```

3. Check Prometheus targets:
   - Open http://localhost:9090/targets
   - Verify target is "up"

### Grafana Dashboard Not Loading

1. Check Grafana is running:
   ```bash
   curl http://localhost:3000
   ```

2. Check Prometheus datasource:
   - Configuration → Data Sources → Prometheus
   - Test connection

3. Check dashboard provisioning:
   - Configuration → Provisioning → Dashboards
   - Verify dashboard is loaded

### CLI Health Endpoint Hanging

The CLI health endpoint may hang if Router service initialization is slow. To fix this, the health command skips Router service initialization:

```go
PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
    return nil
},
PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
    return nil
},
```

## Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [PromQL Reference](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)

## Related Documentation

- `PROMETHEUM_SETUP.md` - Prometheus setup guide
- `GRAFANA_SETUP.md` - Grafana setup guide
- `DOCKER_SETUP.md` - Docker deployment guide
- `CI_CD_SETUP.md` - CI/CD integration
