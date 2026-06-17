# Docker Compose Setup for Ti Router

This directory contains Docker Compose configuration for running Ti Router, Prometheus, and Grafana.

## Quick Start

### Prerequisites

- Docker Desktop installed and running
- Docker Compose v2+

### Start All Services

```bash
docker-compose up -d
```

### Stop All Services

```bash
docker-compose down
```

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f router
docker-compose logs -f prometheus
docker-compose logs -f grafana
```

## Services

### Ti Router
- **Port**: 1807
- **Health Check**: http://localhost:1807/health
- **Metrics**: http://localhost:1807/metrics/prometheus

### Prometheus
- **Port**: 9090
- **Web UI**: http://localhost:9090
- **Config**: ./prometheus.yml
- **Alert Rules**: ./prometheus-alerts.yml

### Grafana
- **Port**: 3000
- **Web UI**: http://localhost:3000
- **Credentials**: admin / admin
- **Dashboard**: Pre-configured Ti Router dashboard

## Volumes

### Data Persistence
- **Router Data**: `./apps/router/data` → `/app/data`
- **Prometheus Data**: Docker volume `prometheus-data`
- **Grafana Data**: Docker volume `grafana-data`

## Building Router Image

```bash
# Build Router image
docker-compose build router

# Rebuild without cache
docker-compose build --no-cache router
```

## Scaling

### Scale Prometheus (for high availability)

```bash
docker-compose up -d --scale prometheus=3
```

Note: This requires additional configuration for Prometheus federation.

## Configuration

### Environment Variables

Router service can be configured with environment variables:

```yaml
environment:
  - PORT=1807
  - LOG_LEVEL=info
  - ENABLE_METRICS=true
```

### Custom Configuration

To use custom configuration files:

1. Create `configs/` directory
2. Add your configuration files
3. Update docker-compose.yml to mount configs

```yaml
volumes:
  - ./configs:/app/configs
```

## Troubleshooting

### Services won't start

```bash
# Check service status
docker-compose ps

# Check logs
docker-compose logs router
```

### Port conflicts

If ports are already in use, modify docker-compose.yml:

```yaml
services:
  router:
    ports:
      - "1808:1807"  # Change to different port
```

### Data not persisting

Ensure data directories exist and have correct permissions:

```bash
mkdir -p apps/router/data
chmod 755 apps/router/data
```

### Prometheus not scraping Router

1. Check Router is running:
   ```bash
   docker-compose ps router
   ```

2. Check Router health:
   ```bash
   curl http://localhost:1807/health
   ```

3. Check Prometheus targets:
   - Open http://localhost:9090/targets
   - Verify Router target is "up"

## Production Deployment

### Use Docker Swarm

```bash
docker stack deploy -c docker-compose.yml ti-stack
```

### Use Kubernetes

Convert docker-compose.yml to Kubernetes manifests using:

```bash
kompose convert
```

### Security Considerations

1. **Change default passwords**:
   - Grafana: Change admin password
   - Router: Add API key authentication

2. **Use secrets management**:
   ```yaml
   secrets:
     router_api_key:
       file: ./secrets/router_api_key.txt
   ```

3. **Enable TLS**:
   - Add reverse proxy (nginx/traefik)
   - Use HTTPS certificates

## Monitoring

### Service Health

```bash
# Check all services
docker-compose ps

# Check specific service
docker-compose exec router curl -f http://localhost:1807/health
```

### Resource Usage

```bash
# Check resource usage
docker stats
```

## Backup and Restore

### Backup Data

```bash
# Backup Router data
docker run --rm -v $(pwd)/apps/router/data:/backup -v $(pwd)/backups:/host alpine tar czf /host/router-backup-$(date +%Y%m%d).tar.gz /backup

# Backup Prometheus data
docker run --rm -v prometheus-data:/data -v $(pwd)/backups:/host alpine tar czf /host/prometheus-backup-$(date +%Y%m%d).tar.gz /data
```

### Restore Data

```bash
# Restore Router data
docker run --rm -v $(pwd)/apps/router/data:/restore -v $(pwd)/backups:/host alpine tar xzf /host/router-backup-YYYYMMDD.tar.gz -C /restore
```

## Resources

- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Dockerfile Best Practices](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)
- [Prometheus Docker Guide](https://prometheus.io/docs/prometheus/latest/installation/)
- [Grafana Docker Guide](https://grafana.com/docs/grafana/latest/installation/docker/)
