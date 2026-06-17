# Grafana Setup for Ti Router

This directory contains the Grafana configuration and setup scripts for visualizing Ti Router metrics.

## Quick Start

### 1. Install Grafana

Run the setup script to download and install Grafana:

```bash
setup-grafana.bat
```

Or manually:

```bash
powershell -ExecutionPolicy Bypass -File setup-grafana.ps1
```

### 2. Start Grafana

Start Grafana with the Ti Router dashboard pre-configured:

```bash
start-grafana.bat
```

Or manually:

```bash
cd Z:\09_TOOLS\grafana
bin\grafana-server.exe
```

### 3. Access Grafana

Open your browser and navigate to: http://localhost:3000

**Default credentials:**
- Username: `admin`
- Password: `admin`

You will be prompted to change the password on first login.

## Dashboard

The Ti Router dashboard includes the following panels:

1. **CLI Registrations** - Rate of CLI registrations over time
2. **Active CLI Instances** - Current number of active CLI instances
3. **CLI Heartbeat Rate** - Rate of heartbeats and failed heartbeats
4. **Router Request Rate** - Rate of incoming requests
5. **Router Error Rate** - Rate of errors
6. **Router Latency P95** - 95th percentile latency
7. **Router Connections** - Active connections
8. **CLI Discoveries** - Rate of CLI discovery operations

## Customization

### Adding New Panels

1. Go to Grafana web UI
2. Navigate to the Ti Router dashboard
3. Click "Add panel" (+)
4. Select panel type
5. Configure Prometheus query
6. Save dashboard

### Custom Queries

Example queries for Ti Router metrics:

```promql
# CLI metrics
rate(ti_router_cli_registrations_total[5m])
ti_router_cli_active_count
rate(ti_router_cli_heartbeats_total[5m])

# Router metrics
rate(ti_router_requests_total[5m])
rate(ti_router_requests_errors[5m])
ti_router_latency_p95_ms
ti_router_connections_active

# Combined metrics
rate(ti_router_cli_heartbeats_total[5m]) / rate(ti_router_cli_registrations_total[5m])
```

## Troubleshooting

### Grafana won't start

1. Check if port 3000 is already in use:
   ```bash
   netstat -ano | findstr :3000
   ```

2. Check Grafana logs:
   ```
   Z:\09_TOOLS\grafana\logs\grafana.log
   ```

### Dashboard not loading

1. Verify Prometheus is running:
   ```bash
   curl http://localhost:9090/api/v1/targets
   ```

2. Check datasource configuration in Grafana:
   - Configuration → Data Sources → Prometheus
   - Verify URL: `http://localhost:9090`
   - Test connection

### No data in dashboard

1. Verify Router is running:
   ```bash
   curl http://localhost:1807/health
   ```

2. Verify metrics endpoint:
   ```bash
   curl http://localhost:1807/metrics/prometheus
   ```

3. Check Prometheus targets:
   - Open http://localhost:9090/targets
   - Verify target is "up"

## Advanced Configuration

### Custom Dashboard

Create a custom dashboard JSON file:

```json
{
  "dashboard": {
    "title": "Custom Dashboard",
    "panels": [
      {
        "id": 1,
        "title": "Custom Panel",
        "type": "graph",
        "targets": [
          {
            "expr": "your_prometheus_query",
            "legendFormat": "Label"
          }
        ]
      }
    ]
  }
}
```

Copy to `Z:\09_TOOLS\grafana\conf\dashboards\` and restart Grafana.

### Multiple Data Sources

Add additional data sources in `conf/provisioning/datasources.yml`:

```yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://localhost:9090
    isDefault: true

  - name: AnotherPrometheus
    type: prometheus
    access: proxy
    url: http://another-prometheus:9090
```

## Security

### Change Default Password

1. Login to Grafana
2. Navigate to Configuration → Users
3. Click on admin user
4. Change password

### Enable Authentication

Edit `conf\grafana.ini`:

```ini
[auth.anonymous]
enabled = false

[auth.basic]
enabled = true
```

## Resources

- [Grafana Documentation](https://grafana.com/docs/)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [PromQL Reference](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Ti Router Metrics Architecture](docs/MONITORING_ARCHITECTURE.md)
