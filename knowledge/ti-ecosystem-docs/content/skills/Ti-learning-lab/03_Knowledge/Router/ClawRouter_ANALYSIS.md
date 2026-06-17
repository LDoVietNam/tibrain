---
title: ClawRouter - Báo Cáo Phân Tích
version: 1.0
created: 2026-03-16
updated: 2026-03-16
status: completed
authors:
  - name: Jarvis
    role: System Orchestrator
tags: [clawrouter, research, analysis, multiagent, llm-router]
related:
  - Z:/AGENTS.md
  - Z:/knowledge_base/AGENTS.md
  - Z:/knowledge_base/taskboard/TASKBOARD_GLOBAL.md
---

# ClawRouter - Báo Cáo Phân Tích

## Metadata

| Thuộc tính | Giá trị |
|-----------|---------|
| **Repo** | https://github.com/BlockRunAI/ClawRouter |
| **Version** | 0.12.44 |
| **License** | MIT |
| **Language** | TypeScript 5.7 |
| **Runtime** | Node.js >= 20 |
| **Ngày phân tích** | 2026-03-16 |
| **Clone location** | Z:/knowledge_base/04_Research/01_Analysis/ClawRouter |

---

## Tổng Quan

### Mục tiêu chính
ClawRouter là **LLM router chuyên cho autonomous agents** với đặc điểm:
- **Agent-native**: Agents không cần signup, credit card, API keys
- **Zero API keys**: Wallet signature = authentication
- **Local routing**: 15-dimension weighted scoring, <1ms latency
- **USDC payments**: Pay per-request qua x402 protocol (Base hoặc Solana)
- **Smart routing**: 41+ models, 92% cost savings

### Điểm khác biệt
Là router duy nhất hỗ trợ **5 tính năng**:
1. Open source
2. Smart routing
3. Runs locally
4. Crypto native
5. Agent ready

---

## Kiến Trúc

### High-level Architecture

```
OpenClaw / Your App (OpenAI-compatible client)
                    ↓
          ClawRouter Proxy (localhost)
    ┌─────────────┬─────────────┬──────────────┐
    │   Dedup     │   Router    │   x402 Payment│
    │   Cache     │  (15-dim)   │  (USDC)       │
    └─────────────┴─────────────┴──────────────┘
    ┌─────────────┬─────────────┬──────────────┐
    │  Fallback   │   Balance   │   SSE Heart- │
    │   Chain     │  Monitor    │   beat        │
    └─────────────┴─────────────┴──────────────┘
                    ↓
          blockrun.ai/api (EVM) ──→ OpenAI/Anthropic/Google
                    ↓
          sol.blockrun.ai/api (Solana) ──→ OpenAI/Anthropic/Google
```

### Request Flow

1. **Request Received**: POST /v1/chat/completions
2. **Deduplication Check**: SHA-256 hash + 30s TTL cache
3. **Smart Routing**: 15-dimension weighted scorer (nếu model = "blockrun/auto")
4. **Balance Check**: Kiểm tra sufficient funds với 1.5x buffer
5. **SSE Heartbeat**: Gửi headers + heartbeat ngay lập tức (streaming)
6. **x402 Payment Flow**:
   - Base (EVM): EIP-712 signing (USDC)
   - Solana: SVM signing (USDC)
7. **Fallback Chain**: Tự động retry với model khác khi provider error

### Tech Stack

| Component | Technology |
|-----------|-----------|
| **Language** | TypeScript 5.7 |
| **Runtime** | Node.js >= 20 |
| **Build** | tsup, vite |
| **Testing** | vitest, tsx |
| **Crypto** | @scure/bip32, @scure/bip39, viem, ethers, @solana/kit |
| **Payment** | @x402/evm, @x402/fetch, @x402/svm |
| **Framework** | OpenClaw plugin (optional) |
| **HTTP Server** | Node.js native (http module) |

---

## Features Chính

### 1. Smart Routing Engine

**15-dimension weighted scorer**:
- Query length
- Token count
- Reasoning required
- Code complexity
- Tool calling
- Vision inputs
- Context size
- Speed requirements
- Quality needs
- Cost sensitivity
- Rate limiting
- Model availability
- Provider reliability
- User preferences
- Session context

**Routing Profiles**:
- `/model auto`: Balanced (74-100% savings) - Default
- `/model eco`: Cheapest possible (95-100% savings)
- `/model premium`: Best quality (0% savings)
- `/model free`: Free tier only (100% savings)

**Tier-based routing**:

| Tier | ECO Model | AUTO Model | PREMIUM Model |
|------|-----------|------------|---------------|
| SIMPLE | nvidia/gpt-oss-120b (FREE) | kimi-k2.5 ($0.60/$3.00) | kimi-k2.5 |
| MEDIUM | gemini-2.5-flash-lite ($0.10/$0.40) | grok-code-fast ($0.20/$1.50) | gpt-5.2-codex ($1.75/$14.00) |
| COMPLEX | gemini-2.5-flash-lite ($0.10/$0.40) | gemini-3.1-pro ($2/$12) | claude-opus-4.6 ($5/$25) |
| REASONING | grok-4-fast ($0.20/$0.50) | grok-4-fast ($0.20/$0.50) | claude-sonnet-4.6 ($3/$15) |

**Blended average**: $2.05/M vs $25/M (Claude Opus) = **92% savings**

### 2. Payment System

**x402 Protocol**:
- **No credit cards**: Pay per-request with USDC
- **Dual-chain support**: Base (EVM) hoặc Solana
- **Non-custodial**: USDC stays in wallet until spent
- **Client-side signing**: Wallet key never leaves machine

**Wallet Management**:
```bash
/wallet              # Check balance and address
/wallet export       # Export private key + mnemonic
/wallet recover      # Restore from mnemonic
/wallet solana       # Switch to Solana USDC
/wallet base         # Switch to Base USDC
```

**Key Derivation**:
- EVM: BIP-32 secp256k1 (m/44'/60'/0'/0/0)
- Solana: SLIP-10 Ed25519 (m/44'/501'/0'/0') - Phantom/Solflare compatible
- Single BIP-39 mnemonic generates both keys

### 3. 41+ Models

**Providers**:
- OpenAI: gpt-5.2, gpt-4o, gpt-4o-mini, gpt-oss-120b (FREE), o1, o1-mini, o3, o4-mini
- Anthropic: claude-opus-4.6, claude-sonnet-4.6, claude-haiku-4.5
- Google: gemini-3.1-pro, gemini-3-pro-preview, gemini-3-flash-preview, gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-lite
- DeepSeek: deepseek-chat, deepseek-reasoner
- xAI: grok-4-0709, grok-4-1-fast-reasoning, grok-code-fast-1
- Moonshot: kimi-k2.5
- MiniMax: minimax-m2.5
- NVIDIA: gpt-oss-120b (FREE)

### 4. Optimizations

**Response Deduplication**:
- SHA-256 hash of request body
- 30s TTL cache
- Prevents double-charging on retries

**SSE Heartbeat**:
- Sends headers + heartbeat immediately
- Prevents 10-15s timeout
- 2s interval

**Response Cache**:
- LRU cache with TTL
- Reduces latency for repeated queries
- Configurable size

**Request Compression**:
- Compresses context > 5MB
- Reduces token usage
- Transparent to clients

**Fallback Chain**:
- Automatic retry on provider errors
- 5 max attempts
- Rate limit tracking (60s cooldown)

### 5. Image Generation

**5 image models**:
- `nano-banana`: Google Gemini Flash ($0.05/image)
- `banana-pro`: Google Gemini Pro ($0.10/image)
- `dall-e-3`: OpenAI DALL-E 3 ($0.04/image)
- `gpt-image`: OpenAI GPT Image 1 ($0.02/image)
- `flux`: Black Forest Flux 1.1 ($0.04/image)

**Commands**:
```
/imagegen a dog dancing on the beach
/imagegen --model dall-e-3 a futuristic city at sunset
/imagegen --model banana-pro --size 2048x2048 mountain landscape
/img2img --image ~/photo.png change the background
```

### 6. Usage Analytics

**Stats tracking**:
- Per-request logging (JSON)
- Daily/monthly aggregates
- Cost savings calculation
- Model usage distribution

**Commands**:
```bash
/stats               # View usage and savings
/stats 7            # Last 7 days
/stats clear        # Reset statistics
```

### 7. AI-Powered Troubleshooting

**Doctor command**:
```bash
npx @blockrun/clawrouter doctor
```

- Collects diagnostics
- Sends to Claude Sonnet for analysis
- Cost: ~$0.003 (Sonnet) or ~$0.01 (Opus)
- Provides actionable fixes

### 8. Partner Integrations

**Built-in tools**:
- OpenClaw integration
- Telegram bot support
- MCP (Model Context Protocol) compatible
- API-first design

---

## Source Structure

```
ClawRouter/
├── src/
│   ├── router/              # Smart routing engine
│   │   ├── index.ts         # Entry point
│   │   ├── config.ts        # Routing config
│   │   ├── strategy.ts      # Router strategy (rules-based)
│   │   ├── selector.ts      # Model selection logic
│   │   ├── rules.ts         # Classification rules
│   │   ├── llm-classifier.ts # LLM-based classifier
│   │   └── types.ts         # TypeScript types
│   ├── compression/         # Context compression
│   ├── partners/            # Partner integrations
│   │   ├── index.ts         # Partner registry
│   │   ├── registry.ts      # Partner definitions
│   │   └── tools.ts         # Partner tools
│   ├── auth.ts              # Wallet & authentication
│   ├── balance.ts           # Balance monitoring
│   ├── wallet.ts            # Key derivation
│   ├── payment-preauth.ts   # Payment pre-authorization
│   ├── proxy.ts             # Main proxy server
│   ├── models.ts            # Model definitions
│   ├── router/              # Smart routing
│   ├── stats.ts             # Usage analytics
│   ├── doctor.ts            # AI troubleshooting
│   ├── dedup.ts             # Request deduplication
│   ├── response-cache.ts    # Response caching
│   ├── session.ts           # Session management
│   ├── journal.ts           # Request journal
│   ├── compression/         # Context compression
│   ├── cli.ts               # CLI commands
│   ├── index.ts             # Plugin entry point
│   ├── config.ts            # Configuration
│   ├── errors.ts            # Error classes
│   ├── logger.ts            # Usage logging
│   ├── retry.ts             # Retry logic
│   ├── solana-balance.ts    # Solana balance check
│   ├── solana-sweep.ts      # Legacy wallet migration
│   ├── spend-control.ts     # Spend control
│   ├── update-hint.ts       # Update notifications
│   ├── updater.ts           # Auto-updater
│   └── version.ts           # Version info
├── docs/
│   ├── architecture.md      # Technical deep-dive
│   ├── configuration.md     # Configuration reference
│   ├── features.md          # Feature documentation
│   ├── image-generation.md  # Image API docs
│   ├── routing-profiles.md  # Routing profiles
│   ├── troubleshooting.md   # Troubleshooting guide
│   └── vs-openrouter.md     # Comparison with OpenRouter
├── skills/                  # AI skills definitions
├── scripts/                 # Utility scripts
├── test/                    # Test files
├── package.json
├── tsconfig.json
└── README.md
```

---

## Multiagent Capabilities

### Agent-Native Design

**Built for agents, not humans**:
- ✅ No account signup required
- ✅ No credit cards needed
- ✅ No API keys to manage
- ✅ Wallet signature = authentication
- ✅ Pay per-request with USDC
- ✅ Local routing (<1ms)
- ✅ Zero external dependencies

**Agent capabilities**:
1. **Autonomous operation**: Agents can generate wallet, sign transactions, pay for requests
2. **No human intervention**: No need for credit cards, accounts, or API keys
3. **Non-custodial**: Full control over wallet and funds
4. **Transparent pricing**: See cost before signing (x402 protocol)
5. **Fallback handling**: Automatic retry on provider errors
6. **Rate limit awareness**: Tracks and avoids rate-limited models

### Partner Integrations

**OpenClaw Plugin**:
- Seamless integration with OpenClaw framework
- Plugin architecture for extensibility
- Auto-configuration injection
- Agent-specific auth profiles

**Telegram Bot Support**:
- Full Telegram integration
- Slash commands for routing
- Image generation via bot
- Wallet management commands

**MCP Compatible**:
- Model Context Protocol support
- Agent-friendly API
- Tool calling support
- Streaming responses

### Skills System

**Built-in AI skills**:
- Image generation
- Image editing (img2img)
- Smart routing
- Cost optimization
- Troubleshooting

---

## So Sánh với CLIProxyAPI (Router Hiện Tại)

### CLIProxyAPI Overview

**Tech Stack**:
- Language: Go (Golang)
- Architecture: Proxy server with multi-provider support
- Providers: OpenAI, Gemini, Claude, Qwen, iFlow, etc.
- Auth: OAuth (Gemini, Claude, Qwen, iFlow), API keys
- Payment: Via upstream providers (credit cards, subscriptions)
- Configuration: YAML-based
- Management: Management API, web dashboard support

**Key Features**:
- Multi-account load balancing
- OAuth authentication flows
- Model mapping & fallback
- Streaming & non-streaming
- Function calling/tools support
- Multimodal input
- Amp CLI integration
- Management API for remote control

### So Sánh Chi Tiết

| Feature | ClawRouter | CLIProxyAPI |
|---------|-----------|-------------|
| **Language** | TypeScript 5.7 | Go |
| **Runtime** | Node.js >= 20 | Go binary |
| **Routing** | 15-dimension smart scoring | Manual config / fill-first strategy |
| **Models** | 41+ models (auto-routed) | Multi-provider (config-based) |
| **Auth** | Wallet signature (x402) | OAuth + API keys |
| **Payment** | USDC per-request (x402) | Credit cards / subscriptions |
| **Agent-ready** | ✅ Yes (no API keys) | ⚠️ Partial (OAuth needs human) |
| **Open Source** | ✅ MIT | ✅ MIT |
| **Local routing** | ✅ Yes (<1ms) | ✅ Yes |
| **Fallback** | ✅ Automatic (5 attempts) | ✅ Configurable |
| **Smart routing** | ✅ 15-dimension scoring | ⚠️ Config-based only |
| **Cost optimization** | ✅ 92% savings average | ⚠️ Manual configuration |
| **Image generation** | ✅ 5 models built-in | ❌ Via providers |
| **Image editing** | ✅ img2img support | ❌ Not available |
| **Usage analytics** | ✅ Built-in stats | ✅ Logging (basic) |
| **Troubleshooting** | ✅ AI-powered (Claude) | ⚠️ Manual / logs |
| **Wallet management** | ✅ Dual-chain (Base/Solana) | ❌ Not applicable |
| **Rate limiting** | ✅ 60s cooldown tracking | ⚠️ Basic retry |
| **Context compression** | ✅ Yes (>5MB) | ❌ Not available |
| **Response caching** | ✅ LRU cache | ❌ Not available |
| **Deduplication** | ✅ 30s TTL | ❌ Not available |
| **SSE heartbeat** | ✅ 2s interval | ⚠️ Provider-dependent |
| **Multi-account** | ❌ Single wallet | ✅ Yes (round-robin) |
| **OAuth support** | ❌ No | ✅ Yes (multiple providers) |
| **Management API** | ❌ Not available | ✅ Yes |
| **Web dashboard** | ❌ Via OpenClaw | ✅ Via plugins |
| **Plugin system** | ✅ OpenClaw plugins | ❌ No |
| **Partner integrations** | ✅ OpenClaw, Telegram | ✅ Amp CLI, IDE extensions |

### Ưu Điểm của ClawRouter

1. **Agent-native**: Được thiết kế riêng cho autonomous agents
2. **Smart routing**: 15-dimension scoring tự động chọn model tối ưu
3. **Cost optimization**: 92% savings so với model cao cấp
4. **USDC payments**: Pay per-request, không cần credit card
5. **Image generation**: 5 models built-in, không cần config
6. **Image editing**: img2img support, tính năng nâng cao
7. **Context compression**: Tối ưu token usage cho context lớn
8. **Response caching**: Giảm latency cho queries lặp lại
9. **Deduplication**: Tránh double-charging khi retry
10. **AI troubleshooting**: Claude-powered doctor command

### Ưu Điểm của CLIProxyAPI

1. **Go performance**: High performance, low memory footprint
2. **Multi-account**: Round-robin load balancing
3. **OAuth support**: Multiple OAuth providers (Gemini, Claude, Qwen, iFlow)
4. **Management API**: Remote control & monitoring
5. **Web dashboard**: Via plugins, rich UI
6. **Mature ecosystem**: Many based projects
7. **Model mapping**: Flexible model mapping & fallback
8. **IDE integration**: Amp CLI, Cursor, Claude Code, etc.
9. **Production-ready**: Battle-tested, stable
10. **Active community**: Many contributors, forks, extensions

### Khi nào nên dùng cái nào?

**Dùng ClawRouter khi**:
- Xây autonomous agents cần hoạt động độc lập
- Muốn tối ưu cost với smart routing tự động
- Cần image generation/editing built-in
- Muốn pay per-request với USDC
- Không muốn quản lý nhiều account
- Cần agent-ready infrastructure

**Dùng CLIProxyAPI khi**:
- Cần multi-account load balancing
- Muốn OAuth authentication
- Cần remote management API
- Xây IDE integrations
- Muốn ecosystem mở rộng (many based projects)
- Cần production-ready Go infrastructure
- Muốn model mapping linh hoạt

---

## Assessment

### Strengths

✅ **Agent-first design**: Được thiết kế từ đầu cho autonomous agents
✅ **Smart routing**: 15-dimension scoring tự động, <1ms latency
✅ **Cost optimization**: 92% savings average
✅ **No API keys**: Wallet signature = authentication
✅ **USDC payments**: Pay per-request, transparent pricing
✅ **Open source**: MIT license
✅ **Dual-chain**: Base (EVM) và Solana support
✅ **Image generation**: 5 models built-in
✅ **Image editing**: img2img support
✅ **Context compression**: Tối ưu token usage
✅ **Response caching**: Giảm latency
✅ **Deduplication**: Tránh double-charging
✅ **AI troubleshooting**: Claude-powered doctor
✅ **Fallback chain**: Automatic retry on errors
✅ **Rate limiting**: 60s cooldown tracking
✅ **Usage analytics**: Built-in stats
✅ **Plugin system**: OpenClaw compatible

### Weaknesses

❌ **No multi-account**: Chỉ support single wallet
❌ **No OAuth**: Không support OAuth authentication
❌ **No management API**: Không có remote control
❌ **Limited ecosystem**: Ít based projects hơn CLIProxyAPI
❌ **Node.js dependency**: Cần Node.js runtime
❌ **Learning curve**: x402 protocol, wallet management mới lạ
❌ **USDC only**: Chỉ chấp nhận USDC, không accept SOL/ETH
❌ **Limited documentation**: Docs chưa đầy đủ như CLIProxyAPI

### Risks

⚠️ **New protocol**: x402 protocol chưa phổ biến, có thể thay đổi
⚠️ **Wallet management**: User phải quản lý mnemonic, private keys
⚠️ **Chain dependency**: Phụ thuộc vào Base/Solana network stability
⚠️ **Provider changes**: BlockRun API có thể thay đổi pricing, availability
⚠️ **Adoption**: Cộng đồng chưa lớn, ít feedback

---

## Recommendations

### 1. Tích hợp Hybrid (Khuyến nghị cao nhất)

**Mô hình kết hợp**:
```
Agent Apps
    ↓
┌─────────────────────────────────┐
│     Hybrid Router Layer         │
│  ┌─────────────┬───────────────┐ │
│  │ ClawRouter  │ CLIProxyAPI   │ │
│  │ (x402)      │ (OAuth/API)   │ │
│  │ Smart route │ Multi-account │ │
│  └─────────────┴───────────────┘ │
└─────────────────────────────────┘
```

**Chiến lược**:
1. **Dùng ClawRouter cho**:
   - Autonomous agents cần độc lập
   - Queries cần cost optimization
   - Image generation/editing
   - Development/testing (USDC cheap)

2. **Dùng CLIProxyAPI cho**:
   - Multi-account load balancing
   - OAuth-based authentication
   - Remote management needs
   - Production deployments
   - IDE integrations

3. **Unified gateway**:
   - Single endpoint cho agents
   - Auto-route dựa trên context
   - Fallback chain giữa 2 routers
   - Unified logging & analytics

**Benefit**:
- Tận dụng ưu điểm của cả 2
- Agent-ready infrastructure (ClawRouter)
- Production-ready management (CLIProxyAPI)
- Flexible payment options
- High availability & resilience

### 2. Học Patterns (Khuyến nghị cao)

**Patterns cần học từ ClawRouter**:

1. **Smart routing engine**:
   - 15-dimension weighted scoring
   - Tier-based model selection
   - Context-aware routing
   - Cost-optimized decisions

2. **Response optimization**:
   - Request deduplication (30s TTL)
   - Response caching (LRU)
   - Context compression (>5MB)
   - SSE heartbeat (2s interval)

3. **Payment abstraction**:
   - x402 protocol (pay per-request)
   - Dual-chain support (EVM/Solana)
   - Transparent pricing (see before sign)
   - Non-custodial wallet

4. **Image generation**:
   - Built-in 5 models
   - img2img support
   - Size flexibility
   - Model switching

5. **Troubleshooting automation**:
   - AI-powered diagnostics
   - Cost-effective (~$0.003)
   - Actionable recommendations
   - Opus mode for complex issues

6. **Rate limit awareness**:
   - Track rate-limited models
   - 60s cooldown
   - Deprioritize affected models
   - Automatic recovery

**Áp dụng vào CLIProxyAPI**:
```yaml
# config.yaml
routing:
  strategy: "smart"  # Thêm smart routing (learning from ClawRouter)
  scoring:
    dimensions: 15   # 15-dimension weighted scorer
    tier-selection: true
    cost-optimization: true
    fallback-chain: true
    max-attempts: 5
    rate-limit-tracking: true
    cooldown-seconds: 60

optimizations:
  request-dedup:
    enabled: true
    ttl-seconds: 30
    algorithm: sha256
  
  response-cache:
    enabled: true
    size: 1000
    ttl-seconds: 60
  
  context-compression:
    enabled: true
    threshold-mb: 5
  
  sse-heartbeat:
    enabled: true
    interval-ms: 2000
```

### 3. Reference Only (Khuyến nghị thấp)

**Khi nên dùng reference-only**:
- Đã có router system ổn định
- Không muốn thay đổi stack
- Chỉ cần tham khảo patterns
- Team chưa quen với x402/USDC

**Cách reference**:
1. Đọc architecture docs
2. Study routing logic
3. Analyze payment flow
4. Adapt patterns locally
5. Không tích hợp code

---

## Implementation Roadmap (nếu tích hợp)

### Phase 1: Evaluation (1-2 weeks)
- [ ] Review codebase thoroughly
- [ ] Benchmark performance
- [ ] Test x402 integration
- [ ] Evaluate USDC onboarding
- [ ] Document findings

### Phase 2: Hybrid Gateway (3-4 weeks)
- [ ] Design unified routing layer
- [ ] Implement context-based routing
- [ ] Integrate ClawRouter SDK
- [ ] Connect existing CLIProxyAPI
- [ ] Add fallback logic
- [ ] Test end-to-end

### Phase 3: Feature Integration (2-3 weeks)
- [ ] Port smart routing to CLIProxyAPI
- [ ] Add request deduplication
- [ ] Implement response caching
- [ ] Add context compression
- [ ] Integrate image generation
- [ ] Enable img2img

### Phase 4: Testing & Documentation (1-2 weeks)
- [ ] Comprehensive testing
- [ ] Performance benchmarking
- [ ] Cost analysis
- [ ] User docs
- [ ] API docs
- [ ] Migration guide

### Phase 5: Deployment (1 week)
- [ ] Canary deployment
- [ ] Monitoring setup
- [ ] Incident response
- [ ] User communication
- [ ] Rollback plan

---

## Kết Luận

ClawRouter là một **LLM router agent-native** với kiến trúc tiên tiến, đặc biệt phù hợp cho autonomous agents. Điểm mạnh nhất là **smart routing** (15-dimension scoring) và **USDC payments** (pay per-request, transparent pricing).

**So với CLIProxyAPI**:
- ClawRouter: Agent-ready, smart routing, cost optimization, image generation
- CLIProxyAPI: Multi-account, OAuth, management API, mature ecosystem

**Khuyến nghị**: **Tích hợp Hybrid** - kết hợp ưu điểm của cả 2:
- Dùng ClawRouter cho autonomous agents, cost optimization, image generation
- Dùng CLIProxyAPI cho multi-account, OAuth, management, production
- Unified gateway với auto-routing và fallback chain

**Patterns cần học**:
- Smart routing engine (15-dimension scoring)
- Response optimization (dedup, cache, compression)
- Payment abstraction (x402, dual-chain)
- Image generation/editing
- AI troubleshooting
- Rate limit awareness

**Risks**: x402 protocol mới, wallet management, chain dependency, provider changes

**Đánh giá tổng thể**: ⭐⭐⭐⭐ (4/5) - Excellent for agents, good for production use with hybrid integration

---

## References

- **ClawRouter GitHub**: https://github.com/BlockRunAI/ClawRouter
- **ClawRouter Docs**: https://blockrun.ai/docs
- **x402 Protocol**: https://x402.org
- **CLIProxyAPI**: Z:/Router/CLIProxyAPI-mainline
- **CLIProxyAPI Docs**: https://help.router-for.me/

---

*Báo cáo được tạo bởi Jarvis - System Orchestrator*
*Ngày tạo: 2026-03-16*
*Version: 1.0*