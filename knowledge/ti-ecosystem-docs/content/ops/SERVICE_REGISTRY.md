# Ti Ecosystem Service Registry

**Version:** 1.0.0
**Last Updated:** 2026-07-09
**Schema:** `json: {port, service, layer, purpose, health, config, dependencies}`
**Purpose:** Single source of truth for all services, ports, and responsibilities

---

## 🏗️ Architecture Layers

```
┌─────────────────────────────────────────────────────────┐
│                    External Clients                       │
│  (Claude Code, Codex, OpenCode, VS Code Extensions)      │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│  Gateway Layer (Tirouter - Port 1807)                   │
│  • AI model routing & load balancing                     │
│  • Protocol translation (Anthropic ↔ OpenAI)           │
│  • Auth management                                       │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│  Control-Plane Layer (TiBrain - Port 1810)                │
│  • Knowledge orchestration                               │
│  • Health & discovery APIs                               │
│  • Cross-brain coordination                              │
└─────────────────────────────────────────────────────────┘
```

---

## Network Ports Map

| Port | Service | Project | Layer | Purpose | Health Endpoint |
|------|---------|---------|-------|---------|-----------------|
| **1807** | Tirouter Gateway | `apps/Tirouter` | Gateway | AI proxy, provider aggregation, protocol translation | `/v1/health` |
| **1810** | TiBrain | `apps/tibrain` | Control-Plane | Knowledge routing, health, discovery APIs | `/health`, `/ready` |
| **1811** | Beads | `apps/tibrain` | API | Notification/emission system | `/emit/health` |
| **3456** | claude-nim | `apps/Tirouter/claude-nim` | Provider | NVIDIA NIM Adapter (Anthropic → OpenAI) | `/v1/models` |

---

## Service Definitions (Machine-Readable)

```yaml
services:
  - name: tirouter-gateway
    port: 1807
    path: apps/Tirouter
    layer: gateway
    entrypoint: CLIProxyAPI/cmd/server
    config: CLIProxyAPI/config.yaml
    health_endpoint: /v1/health
    dependencies:
      - claude-nim:3456
    commands:
      start: go run ./cmd/server
      build: go build -o tirouter.exe ./cmd/server

  - name: tibrain-control
    port: 1810
    path: apps/tibrain
    layer: control-plane
    entrypoint: tibrain.exe
    config: .ti/config.json
    health_endpoint: /health
    dependencies: []

  - name: claude-nim-proxy
    port: 3456
    path: apps/Tirouter/claude-nim
    layer: provider
    entrypoint: bun scripts/server.ts
    config: .vscode/settings.json
    health_endpoint: /v1/models
    dependencies:
      - nvidia-api: https://integrate.api.nvidia.com/v1
```

---

## 📋 Service Registry

### Control-Plane Services (TiBrain)
| Service | Path | Role | Status | Owner |
|---------|------|------|--------|-------|
| TiBrain Central | `apps/tibrain/` | Knowledge routing, health APIs | Active | Core |
| Beads | `apps/tibrain/` | Notification/emission system | Active | Core |

### Gateway Services (Tirouter)
| Service | Path | Role | Status | Owner |
|---------|------|------|--------|-------|
| OmniRoute | `apps/Tirouter/OmniRoute-main` | Multi-provider gateway, dashboard | Active | Gateway |
| CLIProxyAPI | `apps/Tirouter/CLIProxyAPI` | Auth, translation, routing layer | Active | Gateway |
| claude-nim | `apps/Tirouter/claude-nim` | NVIDIA NIM proxy (VS Code) | Active | Provider |

---

## 🔌 Integration Contracts

### Protocol Translation Matrix
| From Format | To Format | Endpoint | Notes |
|-----------|---------|----------|-------|
| Anthropic | OpenAI | `POST /v1/messages` → `POST /v1/chat/completions` | claude-nim handles |
| OpenAI | Provider | `POST /v1/chat/completions` | CLIProxyAPI routes |

### Auth Flow
```
Client → API Key (Tirouter) → Route to Provider → Upstream API
                        ↓
                   OAuth passthrough
                        ↓
                    User creds (NVIDIA, OpenAI, etc.)
```

---

## 📦 Provider Catalog (via Tirouter)

| Alias | Model | Endpoint | Type |
|-------|-------|----------|------|
| `nim-sonnet` | claude-3-5-sonnet-20241022 | http://127.0.0.1:3456 | Claude → OpenAI proxy |
| `nim-deepseek-r1` | deepseek-ai/deepseek-r1 | https://integrate.api.nvidia.com/v1 | NVIDIA NIM direct |
| `nim-llama-3-3-70b` | meta/llama-3.3-70b-instruct | https://integrate.api.nvidia.com/v1 | NVIDIA NIM direct |
| `openai/*` | gpt-4, gpt-3.5-turbo | https://api.openai.com/v1 | OpenAI |
| `groq/*` | llama-3.3-70b, mixtral | https://api.groq.com/openai/v1 | Groq |
| `iflow/*` | qwen3-coder-plus, deepseek-v3.2 | https://apis.iflow.cn/v1 | iFlow |

---

## 🛠️ Operations Guide

### Start Services
```bash
# Control-Plane
cd apps/tibrain && ./tibrain.exe

# Gateway + Providers
cd apps/Tirouter/CLIProxyAPI && go run ./cmd/server
cd apps/Tirouter/OmniRoute-main && npm run dev
# [VS Code] Extensions → claude-nim → Start Proxy
```

### Health Checks
```bash
# Control-Plane
curl http://localhost:1810/health
curl http://localhost:1810/ready

# Gateway
curl http://localhost:1807/v1/health
curl http://localhost:1807/v1/models

# Provider
curl http://localhost:3456/v1/models
```

### Verification Pattern
```bash
# 1. Verify port availability
netstat -an | grep -E "(1807|1810|3456)"

# 2. Health check sequence
for port in 1810 1807 3456; do
  curl -sf "http://localhost:$port/health" && echo "OK: $port" || echo "FAIL: $port"
done
```

---

## 🔄 Cross-Service Boundaries

| Boundary | Tirouter Side | TiBrain Side | Integration Method |
|----------|-------------|-------------|------------------|
| Auth | API keys/OAuth credentials | Auth orchestration | Smart routing via `--query` |
| Routing | Port-based (`1807`) | Port-based (`1810`) | Non-overlapping ports |
| Knowledge | Proxy state | Hub DB (`hub.db`) | Shared via `.ti/` context |

---

## 📊 Service Topology Diagram

```
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐
│   Client    │────▶│  Tirouter    │────▶│ Provider APIs   │
│ (ClaudeCO)  │     │  (1807)      │     │ (NVIDIA/OpenAI) │
└─────────────┘     └───────┬──────┘     └─────────────────┘
                          │
                          ▼
                   ┌──────────────┐
                   │ TiBrain      │
                   │  (1810)      │
                   └──────────────┘
```

---

*Registry maintained by Ti Ecosystem Orchestrator. See `AGENTS.md` for layer-specific guidance.*