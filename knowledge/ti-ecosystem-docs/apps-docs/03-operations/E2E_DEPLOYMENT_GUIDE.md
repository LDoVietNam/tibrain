# End-to-End Deployment Guide

> **Version**: 1.0.0
> **Last Updated**: 2026-05-03
> **Purpose**: Guide for deploying and testing Ti Router and CLI end-to-end

## Prerequisites

- Go 1.21+ installed
- curl installed (for testing endpoints)
- Windows or Linux environment

## Quick Start

### Windows (MINGW64/Git Bash)

```bash
cd Z:\10_WORKPLACE\Ti
bash scripts/deploy-e2e-test.sh
```

### Windows (CMD)

```cmd
cd Z:\10_WORKPLACE\Ti
scripts\deploy-e2e-test.bat
```

### Linux/Mac

```bash
cd /path/to/Ti
bash scripts/deploy-e2e-test.sh
```

## Manual Deployment Steps

### 1. Build Router

```bash
cd apps/router
go build -o routerd.exe ./cmd/routerd
cd ../..
```

### 2. Build CLI

```bash
cd apps/cli
go build -o ti.exe .
cd ../..
```

### 3. Start Router

```bash
cd apps/router/cmd/routerd
./routerd.exe -port 1807
```

### 4. Test Router Health

```bash
curl http://localhost:1807/health
```

Expected response:
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime_seconds": 10.5,
  "active_cli_count": 0,
  "total_requests": 0,
  "error_rate": 0.0,
  "providers_healthy": 0,
  "providers_total": 0
}
```

### 5. Test Router Metrics

```bash
curl http://localhost:1807/metrics/prometheus
```

Expected response: Prometheus text format metrics

### 6. Register CLI Instance

```bash
curl -X POST http://localhost:1807/v1/cli/register \
  -H "Content-Type: application/json" \
  -d '{"id":"test-cli-1","address":"localhost:8081"}'
```

Expected response:
```json
{"status":"registered"}
```

### 7. Send CLI Heartbeat

```bash
curl -X POST http://localhost:1807/v1/cli/heartbeat \
  -H "Content-Type: application/json" \
  -d '{"id":"test-cli-1"}'
```

Expected response:
```json
{"status":"ok"}
```

### 8. Discover CLI Instances

```bash
curl http://localhost:1807/v1/cli/discover
```

Expected response:
```json
[
  {
    "id": "test-cli-1",
    "address": "localhost:8081",
    "last_seen": "2026-05-03T17:00:00Z"
  }
]
```

### 9. Test CLI Health Command

```bash
cd apps/cli
./ti health
```

Expected response:
```json
{
  "status": "healthy",
  "cli_id": "cli-instance",
  "uptime_seconds": 0.5,
  "nats_connected": false,
  "router_connected": true,
  "current_mode": "normal",
  "active_sessions": 0,
  "fallback_mode": false
}
```

## Testing Scenarios

### Scenario 1: Router Startup

1. Start Router
2. Verify health endpoint returns healthy
3. Verify metrics endpoint returns data
4. Check logs for errors

### Scenario 2: CLI Registration

1. Start Router
2. Register CLI instance
3. Verify registration successful
4. Check CLI count in metrics

### Scenario 3: CLI Heartbeat

1. Register CLI instance
2. Send heartbeat
3. Verify heartbeat successful
4. Check heartbeat metrics

### Scenario 4: CLI Discovery

1. Register multiple CLI instances
2. Query discovery endpoint
3. Verify all instances returned
4. Check discovery metrics

### Scenario 5: Health Checks

1. Run Router health check
2. Run CLI health command
3. Verify connection status
4. Check error rates

## Troubleshooting

### Router Won't Start

1. Check if port 1807 is already in use
2. Check Router logs: `/tmp/ti-router.log` (Linux/Mac) or `C:\Users\MIN\AppData\Local\Temp\ti-router.log` (Windows)
3. Verify Router binary exists and is executable
4. Check for missing dependencies

### CLI Registration Fails

1. Verify Router is running
2. Check Router health endpoint
3. Verify correct Router URL and port
4. Check network connectivity

### Metrics Endpoint Not Working

1. Verify Router is running
2. Check if metrics package is imported
3. Verify Prometheus handler is registered
4. Check Router logs for errors

### Health Check Returns Unhealthy

1. Check if providers are configured
2. Check provider health status
3. Verify error rate is below threshold
4. Check Router uptime

## Cleanup

### Stop Services

**Windows:**
```cmd
taskkill /F /IM routerd.exe
taskkill /F /IM ti.exe
```

**Linux/Mac:**
```bash
pkill -f routerd
pkill -f ti
```

Or use the cleanup function in the deployment script (Ctrl+C).

## Next Steps

After successful end-to-end testing:

1. **Setup Prometheus**: Configure Prometheus server to scrape metrics
2. **Setup Grafana**: Create dashboards for monitoring
3. **Configure Alerting**: Set up Alertmanager for alerts
4. **CI/CD Integration**: Add integration tests to CI/CD pipeline
