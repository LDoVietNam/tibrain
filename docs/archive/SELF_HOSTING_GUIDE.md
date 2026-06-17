# Self-Hosting Guide - Ti Brain + Obsidian Integration

## Overview

Hệ thống Ti Brain + Obsidian Integration **hoàn toàn có thể self-host** trên infrastructure của bạn, không phụ thuộc vào cloud services (trừ một số optional features).

## 🏗️ Architecture Self-Hosted

```
Your Server (VPS/On-Premise)
├── Ti Brain (Go)
│   ├── REST API Server
│   ├── RAG System
│   ├── SQLite Database
│   └── Vector Store
├── Obsidian MCP Server (TypeScript/Node.js)
│   ├── MCP Tools (14)
│   ├── MCP Resources (3)
│   └── HTTP/STDIO Transport
├── Obsidian Headless (Node.js)
│   ├── Sync CLI
│   └── Vault Management
├── Python Sync Script
│   ├── Frontmatter Mapping
│   ├── File Watching
│   └── Taxonomy Validation
└── Optional: Obsidian Desktop App
    └── Local REST API Plugin
```

## ✅ Self-Hostable Components

### 1. Ti Brain Core (Go) - 100% Self-Hosted
- ✅ REST API Server
- ✅ RAG System (local vector store)
- ✅ SQLite Database (file-based)
- ✅ Memory Management
- ✅ Agent Orchestration
- ✅ Cross-Brain Communication

### 2. Obsidian MCP Server - 100% Self-Hosted
- ✅ MCP Protocol Implementation
- ✅ 14 MCP Tools
- ✅ 3 MCP Resources
- ✅ HTTP/STDIO Transport
- ✅ Local REST API Client

### 3. Python Sync Script - 100% Self-Hosted
- ✅ File System Operations
- ✅ Frontmatter Mapping
- ✅ Taxonomy Validation
- ✅ File Watching
- ✅ Directory Sync

### 4. Database & Storage - 100% Self-Hosted
- ✅ SQLite Database (local file)
- ✅ File System Storage
- ✅ Markdown Files
- ✅ Vector Index (local)

### 5. Frontmatter Mapper (Go) - 100% Self-Hosted
- ✅ Tag Normalization
- ✅ Scope Validation
- ✅ Bidirectional Mapping
- ✅ Field Conversion

## ⚠️ Dependencies Bên Ngoài

### Required (Không thể self-host):
1. **Obsidian Sync Account** (cho obsidian-headless)
   - Là dịch vụ cloud của Obsidian
   - Cần cho sync vault giữa devices
   - **Workaround**: Chỉ dùng local vault, không dùng Obsidian Sync

2. **AI Model Providers** (cho RAG/Agents)
   - OpenAI, Claude, Gemini, etc.
   - Cần API keys từ providers
   - **Workaround**: Dùng local LLM models (Ollama, LocalAI)

### Optional (Có thể thay thế):
1. **Obsidian Desktop App**
   - Cần cho Local REST API
   - **Workaround**: Chỉ dùng MCP server, không cần desktop app

2. **NPM Packages** (development dependencies)
   - Download từ npm registry
   - **Workaround**: Mirror npm registry nội bộ

3. **Python Packages** (PyPI dependencies)
   - Download từ PyPI
   - **Workaround**: Mirror PyPI nội bộ

## 🚀 Self-Hosting Deployment

### Option 1: VPS/Cloud Server

#### Requirements:
- **OS**: Ubuntu 22.04+ / Debian 12+ / Windows Server
- **CPU**: 2 cores minimum
- **RAM**: 4GB minimum (8GB recommended)
- **Storage**: 20GB+ SSD
- **Network**: Static IP recommended

#### Deployment Steps:

**1. Prepare Server:**
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y nodejs npm python3 python3-pip golang git

# Windows Server
# Download Node.js from nodejs.org
# Download Go from golang.org
# Download Python from python.org
```

**2. Clone Repository:**
```bash
git clone <your-repo-url> /opt/tibrain
cd /opt/tibrain
```

**3. Setup Environment:**
```bash
# Run setup script
chmod +x setup_obsidian_integration.sh
./setup_obsidian_integration.sh

# Or manual setup
cd obsidian-mcp-server
npm install
npm run build

cd ../obsidian-headless
npm install
```

**4. Configure Services:**

**Ti Brain Service (systemd):**
```ini
# /etc/systemd/system/tibrain.service
[Unit]
Description=Ti Brain Knowledge System
After=network.target

[Service]
Type=simple
User=tibrain
WorkingDirectory=/opt/tibrain
Environment="PATH=/usr/local/bin:/usr/bin:/bin"
ExecStart=/opt/tibrain/tibrain
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

**Obsidian MCP Server Service:**
```ini
# /etc/systemd/system/obsidian-mcp.service
[Unit]
Description=Obsidian MCP Server
After=network.target

[Service]
Type=simple
User=tibrain
WorkingDirectory=/opt/tibrain/obsidian-mcp-server
Environment="PATH=/usr/local/bin:/usr/bin:/bin"
EnvironmentFile=/opt/tibrain/.env
ExecStart=/usr/bin/npm run start:http
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

**Sync Service:**
```ini
# /etc/systemd/system/obsidian-sync.service
[Unit]
Description=Obsidian Sync Service
After=network.target

[Service]
Type=simple
User=tibrain
WorkingDirectory=/opt/tibrain
Environment="PATH=/usr/local/bin:/usr/bin:/bin"
EnvironmentFile=/opt/tibrain/.env
ExecStart=/usr/bin/uv run --with pyyaml --with watchdog python sync_obsidian.py \
    --direction obsidian-to-tibrain \
    --vault /opt/tibrain/vault \
    --tibrain /opt/tibrain \
    --pattern "*.md" \
    --watch
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

**5. Enable Services:**
```bash
sudo systemctl enable tibrain
sudo systemctl enable obsidian-mcp
sudo systemctl enable obsidian-sync

sudo systemctl start tibrain
sudo systemctl start obsidian-mcp
sudo systemctl start obsidian-sync
```

**6. Configure Reverse Proxy (Nginx):**
```nginx
# /etc/nginx/sites-available/tibrain
server {
    listen 80;
    server_name your-domain.com;

    location /api/ {
        proxy_pass http://localhost:1810;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /mcp/ {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

**7. Enable SSL (Let's Encrypt):**
```bash
sudo certbot --nginx -d your-domain.com
```

### Option 2: Docker Deployment

**Dockerfile for Ti Brain:**
```dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o tibrain

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/tibrain .
COPY --from=builder /app/knowledge ./knowledge
COPY --from=builder /app/tibrain.db ./tibrain.db

EXPOSE 1810
CMD ["./tibrain"]
```

**Dockerfile for Obsidian MCP Server:**
```dockerfile
FROM node:22-alpine

WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

COPY . .
RUN npm run build

ENV OBSIDIAN_BASE_URL=http://host.docker.internal:27123
ENV OBSIDIAN_VERIFY_SSL=false

EXPOSE 3000
CMD ["npm", "run", "start:http"]
```

**docker-compose.yml:**
```yaml
version: '3.8'

services:
  tibrain:
    build: .
    ports:
      - "1810:1810"
    volumes:
      - ./knowledge:/root/knowledge
      - ./tibrain.db:/root/tibrain.db
    environment:
      - TIBRAIN_DB_PATH=/root/tibrain.db
      - TIBRAIN_KNOWLEDGE_PATH=/root/knowledge
    restart: unless-stopped

  obsidian-mcp:
    build: ./obsidian-mcp-server
    ports:
      - "3000:3000"
    environment:
      - OBSIDIAN_API_KEY=${OBSIDIAN_API_KEY}
      - OBSIDIAN_BASE_URL=http://host.docker.internal:27123
    extra_hosts:
      - "host.docker.internal:host-gateway"
    restart: unless-stopped

  obsidian-sync:
    build: .
    command: >
      sh -c "apk add py3-pip && pip3 install pyyaml watchdog &&
             python3 sync_obsidian.py --direction obsidian-to-tibrain
             --vault /vault --tibrain /app --pattern '*.md' --watch"
    volumes:
      - ./vault:/vault
      - ./knowledge:/app/knowledge
    environment:
      - VAULT_PATH=/vault
      - TIBRAIN_PATH=/app
    restart: unless-stopped
```

**Run Docker:**
```bash
docker-compose up -d
docker-compose logs -f
```

### Option 3: Kubernetes Deployment

**ConfigMap:**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: tibrain-config
data:
  OBSIDIAN_API_KEY: "your-api-key"
  OBSIDIAN_BASE_URL: "http://127.0.0.1:27123"
  VAULT_PATH: "/vault"
  TIBRAIN_PATH: "/tibrain"
```

**Deployment:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tibrain
spec:
  replicas: 1
  selector:
    matchLabels:
      app: tibrain
  template:
    metadata:
      labels:
        app: tibrain
    spec:
      containers:
      - name: tibrain
        image: tibrain:latest
        ports:
        - containerPort: 1810
        volumeMounts:
        - name: knowledge
          mountPath: /knowledge
        - name: config
          mountPath: /etc/config
        envFrom:
        - configMapRef:
            name: tibrain-config
      volumes:
      - name: knowledge
        persistentVolumeClaim:
          claimName: knowledge-pvc
      - name: config
        configMap:
          name: tibrain-config
```

**Service:**
```yaml
apiVersion: v1
kind: Service
metadata:
  name: tibrain
spec:
  selector:
    app: tibrain
  ports:
  - protocol: TCP
    port: 80
    targetPort: 1810
  type: LoadBalancer
```

## 🔒 Security Considerations

### 1. Network Security
- **Firewall**: Chỉ mở ports cần thiết (1810, 3000)
- **VPN**: Sử dụng VPN cho remote access
- **SSL/TLS**: Bắt buộc HTTPS cho production

### 2. Authentication
- **API Keys**: Không commit vào git
- **Environment Variables**: Sử dụng secret management
- **RBAC**: Implement role-based access control

### 3. Data Protection
- **Backups**: Regular backups của database và vault
- **Encryption**: Encrypt sensitive data at rest
- **Access Logs**: Monitor và audit access

### 4. Update Management
- **Patching**: Regular OS and dependency updates
- **Scanning**: Vulnerability scanning定期
- **Testing**: Test updates ở staging trước production

## 💾 Backup Strategy

### 1. Database Backup
```bash
# SQLite backup
cp tibrain.db backups/tibrain-$(date +%Y%m%d).db

# Automated backup (cron)
0 2 * * * cp /opt/tibrain/tibrain.db /backups/tibrain-$(date +\%Y\%m\%d).db
```

### 2. Vault Backup
```bash
# Vault directory backup
rsync -av /opt/tibrain/vault/ /backups/vault-$(date +%Y%m%d)/

# Automated backup
0 3 * * * rsync -av /opt/tibrain/vault/ /backups/vault-$(date +\%Y\%m\%d)/
```

### 3. Knowledge Base Backup
```bash
# Knowledge directory backup
rsync -av /opt/tibrain/knowledge/ /backups/knowledge-$(date +%Y%m%d)/
```

### 4. Configuration Backup
```bash
# Environment variables and configs
cp .env backups/.env-$(date +%Y%m%d)
cp TAGS_SIMPLE.md backups/TAGS_SIMPLE.md-$(date +%Y%m%d)
cp SCOPES_SIMPLE.md backups/SCOPES_SIMPLE.md-$(date +%Y%m%d)
```

## 📊 Monitoring

### 1. Health Checks
```bash
# Ti Brain API
curl http://localhost:1810/api/health

# Obsidian MCP
curl http://localhost:3000/api/status

# Service status
systemctl status tibrain
systemctl status obsidian-mcp
systemctl status obsidian-sync
```

### 2. Logs
```bash
# Ti Brain logs
journalctl -u tibrain -f

# MCP Server logs
journalctl -u obsidian-mcp -f

# Sync logs
journalctl -u obsidian-sync -f
```

### 3. Metrics (Optional)
- **Prometheus**: Export metrics
- **Grafana**: Visualize metrics
- **AlertManager**: Alert on failures

## 🔧 Maintenance

### 1. Updates
```bash
# Update Ti Brain
cd /opt/tibrain
git pull
go build
systemctl restart tibrain

# Update MCP Server
cd /opt/tibrain/obsidian-mcp-server
git pull
npm install
npm run build
systemctl restart obsidian-mcp

# Update Sync Script
cd /opt/tibrain
git pull
uv sync
systemctl restart obsidian-sync
```

### 2. Database Maintenance
```bash
# SQLite vacuum
sqlite3 tibrain.db "VACUUM;"

# Database integrity check
sqlite3 tibrain.db "PRAGMA integrity_check;"
```

### 3. Log Rotation
```bash
# Configure logrotate
# /etc/logrotate.d/tibrain
/opt/tibrain/logs/*.log {
    daily
    rotate 7
    compress
    missingok
    notifempty
}
```

## 💰 Cost Estimation

### Self-Hosting Costs (Monthly):

**VPS Option:**
- VPS (2CPU, 4GB RAM): $10-20/month
- Domain: $1/month
- SSL: Free (Let's Encrypt)
- Backup Storage: $5/month
- **Total**: ~$15-25/month

**On-Premise:**
- Hardware: One-time cost
- Electricity: $10-20/month
- Maintenance: Time cost
- **Total**: ~$10-20/month + hardware

### Cloud Services (Comparison):
- Obsidian Sync: $10/month (optional)
- AI Provider APIs: $5-50/month (usage-based)
- Database hosting: $15-50/month
- **Total**: ~$30-110/month

**Savings**: Self-hosting saves ~$15-85/month

## 🎯 Best Practices

### 1. Infrastructure
- Sử dụng configuration management (Ansible, Chef)
- Implement infrastructure as code (Terraform)
- Use containerization (Docker, Kubernetes)

### 2. Operations
- Monitor system metrics
- Set up alerting
- Document runbooks
- Regular testing of disaster recovery

### 3. Security
- Regular security audits
- Keep dependencies updated
- Use strong authentication
- Implement network segmentation

### 4. Performance
- Monitor resource usage
- Optimize database queries
- Cache frequently accessed data
- Scale horizontally when needed

## 🐛 Troubleshooting

### 1. Service Won't Start
```bash
# Check logs
journalctl -u tibrain -n 50

# Check port conflicts
netstat -tulpn | grep 1810

# Check file permissions
ls -la /opt/tibrain
```

### 2. Database Errors
```bash
# Check database integrity
sqlite3 tibrain.db "PRAGMA integrity_check;"

# Recover corrupted database
sqlite3 tibrain.db ".dump" | sqlite3 recovered.db
```

### 3. Sync Failures
```bash
# Test sync manually
python3 sync_obsidian.py --direction obsidian-to-tibrain \
    --vault /vault --tibrain /opt/tibrain --pattern "*.md"

# Check file permissions
ls -la /vault
ls -la /opt/tibrain/knowledge
```

### 4. Memory Issues
```bash
# Check memory usage
free -h
top

# Optimize SQLite
sqlite3 tibrain.db "PRAGMA journal_mode=WAL;"
sqlite3 tibrain.db "PRAGMA synchronous=NORMAL;"
```

## 📋 Self-Hosting Checklist

### Pre-Deployment:
- [ ] Server provisioning completed
- [ ] Operating system updated
- [ ] Firewall configured
- [ ] DNS configured
- [ ] SSL certificate obtained
- [ ] Database backup strategy defined
- [ ] Monitoring setup
- [ ] Alerting configured

### Deployment:
- [ ] Application cloned
- [ ] Dependencies installed
- [ ] Environment configured
- [ ] Services created (systemd/docker)
- [ ] Services started
- [ ] Health checks passing
- [ ] Reverse proxy configured
- [ ] SSL enabled

### Post-Deployment:
- [ ] First sync tested
- [ ] API endpoints tested
- [ ] Monitoring verified
- [ ] Backup procedure tested
- [ ] Documentation updated
- [ ] Team trained
- [ ] Support procedures defined

## Conclusion

**Hệ thống Ti Brain + Obsidian Integration hoàn toàn có thể self-host** với minimal dependencies bên ngoài. Self-hosting mang lại:

✅ **Cost Savings**: $15-85/month tiết kiệm
✅ **Data Privacy**: Full control over data
✅ **Customization**: Complete flexibility
✅ **No Vendor Lock-in**: Independent of cloud services
✅ **Performance**: Local processing, low latency
✅ **Security**: Control over security measures

Chỉ cần một VPS cơ bản hoặc on-premises server là có thể chạy toàn bộ hệ thống.

---

**Last Updated**: 2026-05-23
**Self-Hosting**: ✅ Fully Supported
