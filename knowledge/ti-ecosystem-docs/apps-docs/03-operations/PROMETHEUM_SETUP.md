# Prometheus Setup for Ti Router

This directory contains the Prometheus configuration and setup scripts for monitoring Ti Router metrics.

## Quick Start

### 1. Install Prometheus

Run the setup script to download and install Prometheus:

```bash
setup-prometheus.bat
```

Or manually:

```bash
powershell -ExecutionPolicy Bypass -File setup-prometheus.ps1
```

### 2. Start Prometheus

Start Prometheus with the Ti Router configuration:

```bash
start-prometheus.bat
```

Or manually:

```bash
cd Z:\09_TOOLS\prometheus
prometheus.exe --config.file=prometheus.yml
```

### 3. Access Prometheus Web UI

Open your browser and navigate to: http://localhost:9090

## Configuration

The Prometheus configuration file (`prometheus.yml`) is configured to scrape metrics from:

- **Ti Router**: `http://localhost:1807/metrics/prometheus`
- **Scrape Interval**: 15 seconds

## Available Metrics

### Router Metrics

- `ti_router_requests_total` - Total requests
- `ti_router_requests_success` - Successful requests
- `ti_router_requests_errors` - Failed requests
- `ti_router_request_rate_rps` - Requests per second
- `ti_router_request_rate_rpm` - Requests per minute
- `ti_router_latency_avg_ms` - Average latency
- `ti_router_latency_p50_ms` - P50 latency
- `ti_router_latency_p95_ms` - P95 latency
- `ti_router_latency_p99_ms` - P99 latency
- `ti_router_tokens_prompt` - Total prompt tokens
- `ti_router_tokens_completion` - Total completion tokens
- `ti_router_tokens_total` - Total tokens
- `ti_router_cost_total_usd` - Total cost in USD
- `ti_router_connections_active` - Active connections
- `ti_router_connections_peak` - Peak connections

### CLI Metrics

- `ti_router_cli_registrations_total` - Total CLI registrations
- `ti_router_cli_heartbeats_total` - Total CLI heartbeats
- `ti_router_cli_heartbeats_failed` - Failed CLI heartbeats
- `ti_router_cli_active_count` - Active CLI instances
- `ti_router_cli_discoveries_total` - Total CLI discoveries

## Example Queries

### Router Performance

- Request rate: `rate(ti_router_requests_total[1m])`
- Error rate: `rate(ti_router_requests_errors[1m])`
- Average latency: `ti_router_latency_avg_ms`
- P95 latency: `ti_router_latency_p95_ms`

### CLI Activity

- Active CLI instances: `ti_router_cli_active_count`
- Registration rate: `rate(ti_router_cli_registrations_total[5m])`
- Heartbeat rate: `rate(ti_router_cli_heartbeats_total[5m])`

## Troubleshooting

### Prometheus won't start

1. Check if port 9090 is already in use:
   ```bash
   netstat -ano | findstr :9090
   ```

2. Check the configuration file syntax:
   ```bash
   cd Z:\09_TOOLS\prometheus
   prometheus.exe --config.file=prometheus.yml --dry-run
   ```

### Metrics not appearing

1. Verify Ti Router is running:
   ```bash
   curl http://localhost:1807/health
   ```

2. Verify metrics endpoint is accessible:
   ```bash
   curl http://localhost:1807/metrics/prometheus
   ```

3. Check Prometheus targets page: http://localhost:9090/targets

## Advanced Configuration

To add more targets to scrape, edit `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'ti-router'
    static_configs:
      - targets: ['localhost:1807']
    metrics_path: '/metrics/prometheus'

  - job_name: 'another-service'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

## Documentation

- Prometheus Documentation: https://prometheus.io/docs/
- PromQL (Prometheus Query Language): https://prometheus.io/docs/prometheus/latest/querying/basics/
- Ti Router Metrics Architecture: `docs/MONITORING_ARCHITECTURE.md`
