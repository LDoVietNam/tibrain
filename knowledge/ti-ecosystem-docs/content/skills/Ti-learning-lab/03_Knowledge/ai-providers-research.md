---
tags: ["tibrain", "documentation", "provider", "skill", "authentication"]
scopes: ["auth", "tibrain"]
last_updated: 2026-05-22
---
# AI Providers Research Documentation

> **Purpose**: Document research findings on AI providers similar to Supermaven for future reference and integration.

---

## 📋 Table of Contents

1. [Research Overview](#research-overview)
2. [Provider Analysis](#provider-analysis)
3. [Integration Patterns](#integration-patterns)
4. [Deployment Guide](#deployment-guide)
5. [Repository Locations](#repository-locations)
6. [Lessons Learned](#lessons-learned)

---

## 🎯 Research Overview

### Objective
Find and evaluate AI code completion providers similar to Supermaven that can be integrated into the Ti Router system.

### Research Date
2026-05-06

### Key Findings
- **3 viable providers** identified and cloned
- **Amazon Q Developer** - Real HTTP API, educational license only
- **FauxPilot** - Open source, GitHub Copilot compatible, requires NVIDIA GPU
- **Quack Companion** - Apache 2.0, OSS LLMs, commercial friendly

---

## 📊 Provider Analysis

### 1. Amazon Q Developer (q2api)

**Repository**: https://github.com/CassiopeiaCode/q2api

**Location**: `Z:\10_WORKPLACE\Ti\apps\integrations\providers\amazonq-bridge-provider`

**License**: Educational only (仅供学习和测试使用)

**Tech Stack**: FastAPI, Python, Docker, SQLite/PostgreSQL/MySQL

**API Compatibility**:
- ✅ OpenAI Chat Completions API (`/v1/chat/completions`)
- ✅ Claude Messages API (`/v1/messages`)
- ✅ Tool Use support
- ✅ Streaming support

**Features**:
- Multi-account management
- Auto token refresh
- Web console
- Load balancing
- HTTP proxy support

**Models**: claude-sonnet-4, claude-sonnet-4.5

**Authentication**: AWS OIDC device flow

**Deployment**:
```bash
cd amazonq-bridge-provider
cp .env.example .env
docker-compose up -d
# OR
pip install -r requirements.txt
uvicorn app:app --host 0.0.0.0 --port 8000
```

**Limitations**: Educational use only, not for commercial deployment

---

### 2. FauxPilot

**Repository**: https://github.com/fauxpilot/fauxpilot

**Location**: `Z:\10_WORKPLACE\Ti\apps\integrations\providers\fauxpilot-proxy-provider`

**License**: Open source

**Tech Stack**: Docker, NVIDIA Triton, CodeGen models

**Requirements**: NVIDIA GPU with Compute Capability >= 6.0

**API Compatibility**:
- ✅ GitHub Copilot compatible API (`/v1/completions`)

**Features**:
- Local model inference
- Docker deployment
- Self-hosted

**Models**: CodeGen models (downloaded from Huggingface)

**Authentication**: Self-hosted (no external auth)

**Deployment**:
```bash
cd fauxpilot-proxy-provider
./setup.sh  # Choose and download model
docker-compose up -d
```

**Limitations**: Requires NVIDIA GPU, model download time

---

### 3. Quack Companion

**Repository**: https://github.com/quack-ai/companion

**Location**: `Z:\10_WORKPLACE\Ti\apps\integrations\providers\quack-companion-provider`

**License**: Apache 2.0 (commercial friendly)

**Tech Stack**: FastAPI, Ollama, Python

**API Compatibility**:
- ✅ OpenAI Chat Completions API (`/v1/chat/completions`)

**Features**:
- OSS models support
- Team knowledge integration
- VSCode extension
- Local deployment

**Models**: Mistral, Gemma, Phi 3, Llama 3 (via Ollama)

**Authentication**: API key based

**Deployment**:
```bash
cd quack-companion-provider
# Install Ollama first
ollama pull mistral
ollama pull gemma
# Then deploy
pip install -r requirements.txt
uvicorn app:app --host 0.0.0.0 --port 8000
```

**Limitations**: Requires Ollama installation and model management

---

## 🔧 Integration Patterns

### Common Patterns Identified

1. **OpenAI Compatibility**
   - All providers implement `/v1/chat/completions`
   - Standard request/response format
   - Easy integration with existing router

2. **Authentication Methods**
   - **API Key**: Simple Bearer token authentication
   - **OAuth**: Device flow for cloud providers
   - **Self-hosted**: No authentication required

3. **Deployment Models**
   - **Docker Compose**: Production-ready deployment
   - **Local Python**: Development deployment
   - **GPU Required**: Local inference

4. **Configuration Management**
   - `.env.example` → `.env` pattern
   - Environment variable configuration
   - Port management

### Router Integration Template

```yaml
# providers.yaml template
{provider-name}:
  name: {provider-name}
  base_url: http://localhost:{port}/v1
  api_key_env: ""
  api_keys:
    - {api-key}
  models:
    - {model-1}
    - {model-2}
  format: openai
  timeout_sec: 30
  priority: {priority}
  weight: {weight}
  cost_per_1k: {cost}
```

---

## 🚀 Deployment Guide

### Step-by-Step Process

1. **Provider Selection**
   - Check license requirements
   - Verify system requirements
   - Confirm API compatibility

2. **Environment Setup**
   ```bash
   cd Z:\10_WORKPLACE\Ti\apps\integrations\providers\{provider-name}
   cp .env.example .env
   # Edit .env with required configuration
   ```

3. **Dependency Installation**
   ```bash
   # Python providers
   pip install -r requirements.txt
   
   # Docker providers
   docker-compose up -d
   ```

4. **Service Testing**
   ```bash
   curl -X POST "http://localhost:{port}/v1/chat/completions" \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer {api-key}" \
     -d '{"model": "{model-name}", "messages": [{"role": "user", "content": "test"}]}'
   ```

5. **Router Integration**
   - Add to `providers.yaml`
   - Restart router
   - Test via router endpoint

6. **Validation**
   ```bash
   curl "http://localhost:1807/v1/models" | grep {provider-name}
   ```

### Port Allocation

| Provider | Default Port | Alternative |
|----------|--------------|-------------|
| Amazon Q | 8000 | 8001, 8002 |
| FauxPilot | 5000 | 5001, 5002 |
| Quack Companion | 8000 | 8001, 8002 |

---

## 📁 Repository Locations

### Primary Locations

1. **Providers Folder**: `Z:\10_WORKPLACE\Ti\apps\integrations\providers\`
   - `amazonq-bridge-provider/`
   - `fauxpilot-proxy-provider/`
   - `quack-companion-provider/`

2. **Learning Lab**: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\07_Repositories\`
   - Original repositories for reference
   - Documentation and patterns

3. **Router Patterns**: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\07_Repositories\router\`
   - OmniRoute patterns for advanced routing
   - Executor patterns for provider implementations

### File Structure

```
Z:\10_WORKPLACE\Ti\apps\integrations\providers\
├── README.md                           # Provider documentation
├── notionai-openai-provider/           # Existing provider
├── amazonq-bridge-provider/            # Amazon Q Developer
│   ├── app.py                          # Main FastAPI app
│   ├── docker-compose.yml              # Docker deployment
│   ├── .env.example                    # Environment template
│   └── README.md                       # Provider documentation
├── fauxpilot-proxy-provider/           # FauxPilot
│   ├── setup.sh                        # Model setup script
│   ├── docker-compose.yaml            # Docker deployment
│   └── README.md                       # Provider documentation
└── quack-companion-provider/           # Quack Companion
    ├── docker-compose.yml              # Docker deployment
    ├── pyproject.toml                  # Python dependencies
    └── README.md                       # Provider documentation
```

---

## 📚 Lessons Learned

### Key Insights

1. **License Compliance is Critical**
   - Amazon Q: Educational only
   - FauxPilot: Open source
   - Quack Companion: Apache 2.0 (commercial friendly)

2. **API Standardization Matters**
   - OpenAI compatibility is the de facto standard
   - All providers implement `/v1/chat/completions`
   - Easy integration with existing router

3. **Deployment Complexity Varies**
   - **Simple**: API key + HTTP endpoint
   - **Medium**: Docker + environment setup
   - **Complex**: GPU + model management

4. **System Requirements**
   - **CPU Only**: Amazon Q, Quack Companion
   - **GPU Required**: FauxPilot
   - **External Services**: Ollama (for Quack Companion)

### Best Practices

1. **Always check license first**
2. **Test provider API directly before router integration**
3. **Use consistent naming conventions**
4. **Document port allocations**
5. **Maintain environment templates**
6. **Test with development API keys**

### Common Issues

1. **Port conflicts** - Use different ports for each provider
2. **Authentication failures** - Verify API key format
3. **Router not recognizing provider** - Restart router after config changes
4. **GPU requirements** - Check hardware compatibility

### Future Improvements

1. **Automated provider discovery**
2. **Standardized health checks**
3. **Load balancing across providers**
4. **Automated testing pipeline**
5. **Provider performance monitoring**

---

## 🎯 Recommendations

### For Production Use

1. **Quack Companion** - Apache 2.0, commercial friendly
2. **FauxPilot** - Open source, self-hosted (if GPU available)

### For Development/Testing

1. **Amazon Q Developer** - Educational only, good for learning
2. **All providers** - For pattern research

### For Router Integration

1. **Start with Quack Companion** - Easiest deployment
2. **Add FauxPilot** - If GPU available
3. **Test Amazon Q** - For learning purposes

---

## 📞 Support and References

### Documentation Links
- [Amazon Q Developer](https://github.com/CassiopeiaCode/q2api)
- [FauxPilot](https://github.com/fauxpilot/fauxpilot)
- [Quack Companion](https://github.com/quack-ai/companion)
- [OmniRoute Patterns](https://github.com/diegosouzapw/OmniRoute)

### Internal References
- [Router Configuration](Z:\10_WORKPLACE\Ti\apps\core\router\configs\providers.yaml)
- [Integration Examples](Z:\10_WORKPLACE\Ti\apps\integrations\providers\notionai-openai-provider\)
- [Learning Lab](Z:\10_WORKPLACE\Ti\Ti-learning-lab\)

---

**Last Updated**: 2026-05-06
**Status**: Research completed, providers cloned and documented
**Next Steps**: Deploy and test providers in router environment
