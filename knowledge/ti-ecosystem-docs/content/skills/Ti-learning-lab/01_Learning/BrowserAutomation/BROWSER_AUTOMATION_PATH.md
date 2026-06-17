# Browser Automation Learning Path

> **Focus**: Browser automation, anti-detect fingerprinting, AI-powered browser agents
> **Duration**: 6-8 weeks
> **Tech Stack**: Rust, Python, TypeScript, Go

---

## 🎯 Learning Objectives

Sau khi hoàn thành learning path này, bạn sẽ:
- Hiểu browser fingerprint spoofing techniques
- Master anti-detect browser architecture
- Build AI-powered browser automation agents
- Integrate browser automation với MCP and AI systems
- Implement fingerprint evasion strategies

---

## 📚 Repositories to Study

| # | Repository | Tech Stack | Focus | Duration |
|---|------------|------------|-------|----------|
| 1 | **Donut Browser** | Rust, Tauri, Chromium/Firefox | Anti-detect browser with MCP integration | 2 weeks |
| 2 | **undetectable-fingerprint-browser** | Python, Selenium, Playwright | Fingerprint spoofing and evasion | 1 week |
| 3 | **browser-use** | Python, Playwright, LLM | AI-powered browser automation | 2 weeks |
| 4 | **Skyvern** | Python, Playwright, LangChain | Enterprise browser automation platform | 2 weeks |

---

## 🗓️ Learning Schedule

### Week 1-2: Donut Browser (Rust + Tauri)

**Mục tiêu**: Hiểu anti-detect browser architecture và fingerprint spoofing

#### Week 1: Core Architecture

```
Study Path:
1. Read README.md for features overview
   - Unlimited browser profiles with fingerprint spoofing
   - Chromium & Firefox engines (Wayfern, Camoufox)
   - Proxy/VPN support per profile
   - Local API & MCP server

2. Explore Architecture:
   - src-tauri/ (Tauri Rust backend)
   - src/ (Frontend - likely React/Vue)
   - Study fingerprint spoofing mechanisms
   - Explore proxy/VPN configuration per profile

3. MCP Integration:
   - Study MCP server implementation
   - Explore profile management via MCP
   - Test MCP tools for browser automation

4. Practice:
   - Build: cargo build (Rust)
   - Run: npm run tauri dev (Tauri dev mode)
   - Test profile creation
   - Test proxy configuration
```

**Key Learnings**:
- Rust + Tauri desktop app architecture
- Browser fingerprint spoofing techniques
- Profile isolation and management
- MCP server implementation in Rust
- Proxy/VPN per-profile configuration

#### Week 2: Go Client Integration

```
Study Path:
1. Explore Go Client:
   - Read: apps/cli/internal/donutbrowser/client.go
   - Study Profile Management:
     * CreateProfile(), GetProfile(), UpdateProfile()
     * DeleteProfile(), RunProfile(), KillProfile()
   - Study Proxy Management:
     * CreateProxy(), GetProxy(), UpdateProxy()
     * DeleteProxy(), ListProxies()

2. Integration with Ti CLI:
   - Read: apps/cli/cmd/notion.go
   - Study CLI commands:
     * ti notion register-manual
     * ti notion list-profiles
     * ti notion proxy add/list/delete

3. Practice:
   - Build: cd apps/cli && go build ./...
   - Test: ti notion register-manual --email test@example.com --proxy http://proxy:8080
   - Test: ti notion list-profiles
   - Test: ti notion proxy add --url http://proxy:8080
```

**Key Learnings**:
- Go REST client patterns
- Profile management via API
- Proxy management via API
- CLI command patterns (Cobra)
- Integration patterns (Go client → Rust backend)

#### Capture Learnings to TiBrain

```powershell
# After learning each pattern, capture learning:
.\add-learning-to-tibrain.ps1 `
    -RepoName "donut-browser" `
    -PatternName "rust-tauri-browser-automation" `
    -Description "Rust + Tauri desktop app for anti-detect browser with fingerprint spoofing and MCP integration" `
    -Category "browser" `
    -Family "rust" `
    -Tags "rust,tauri,browser,fingerprint"

.\add-learning-to-tibrain.ps1 `
    -RepoName "donut-browser" `
    -PatternName "fingerprint-spoofing-techniques" `
    -Description "Browser fingerprint spoofing techniques for anti-detection including user agent, screen resolution, canvas fingerprinting" `
    -Category "browser" `
    -Family "security" `
    -Tags "fingerprint,security,anti-detect"
```

---

### Week 3: undetectable-fingerprint-browser (Python)

**Mục tiêu**: Hiểu fingerprint spoofing và evasion techniques

```
Study Path:
1. Read README.md for overview
2. Explore fingerprint spoofing techniques:
   - User agent spoofing
   - Screen resolution spoofing
   - Canvas fingerprinting evasion
   - WebGL fingerprinting evasion
   - Audio fingerprinting evasion
3. Study Selenium/Playwright integration
4. Practice:
   - pip install -r requirements.txt
   - Run examples
   - Test fingerprint evasion
```

**Key Learnings**:
- Fingerprint spoofing techniques
- Selenium/Playwright patterns
- Evasion strategies for detection systems
- Python browser automation

#### Capture Learnings to TiBrain

```powershell
.\add-learning-to-tibrain.ps1 `
    -RepoName "undetectable-fingerprint-browser" `
    -PatternName "fingerprint-evasion-techniques" `
    -Description "Fingerprint evasion techniques including canvas, WebGL, audio fingerprint spoofing" `
    -Category "browser" `
    -Family "python" `
    -Tags "fingerprint,evasion,python,selenium"
```

---

### Week 4-5: browser-use (Python + LLM)

**Mục tiêu**: Hiểu AI-powered browser automation với LLM

```
Study Path:
1. Read README.md for overview
2. Explore architecture:
   - browser_use/ (main library)
   - browser_use/actor/ (LLM actor)
   - browser_use/llm/ (LLM integration)
3. Study LLM integration:
   - How LLM controls browser
   - Natural language to actions
   - Task planning and execution
4. Practice:
   - pip install browser-use
   - Run examples
   - Build simple automation
```

**Key Learnings**:
- LLM-powered browser automation
- Natural language to action translation
- Task planning and execution
- Python Playwright patterns

#### Capture Learnings to TiBrain

```powershell
.\add-learning-to-tibrain.ps1 `
    -RepoName "browser-use" `
    -PatternName "llm-browser-automation" `
    -Description "LLM-powered browser automation with natural language control and Playwright integration" `
    -Category "browser" `
    -Family "ai" `
    -Tags "llm,automation,playwright,python"

.\add-learning-to-tibrain.ps1 `
    -RepoName "browser-use" `
    -PatternName "natural-language-browser-control" `
    -Description "Natural language to browser action translation using LLM" `
    -Category "browser" `
    -Family "ai" `
    -Tags "llm,natural-language,automation"
```

---

### Week 6-7: Skyvern (Python + LangChain)

**Mục tiêu**: Hiểu enterprise browser automation platform

```
Study Path:
1. Read README.md for overview
2. Explore architecture:
   - skyvern-ts/ (TypeScript frontend)
   - skyvern/ (Python backend)
   - integrations/ (LangChain, etc.)
3. Study enterprise features:
   - Multi-user support
   - Task orchestration
   - Integration patterns
4. Practice:
   - docker-compose up
   - Run examples
   - Test automation workflows
```

**Key Learnings**:
- Enterprise browser automation architecture
- LangChain integration patterns
- Task orchestration
- Multi-user systems

#### Capture Learnings to TiBrain

```powershell
.\add-learning-to-tibrain.ps1 `
    -RepoName "skyvern" `
    -PatternName "enterprise-browser-automation" `
    -Description "Enterprise browser automation platform with multi-user support and task orchestration" `
    -Category "browser" `
    -Family "enterprise" `
    -Tags "enterprise,automation,langchain"

.\add-learning-to-tibrain.ps1 `
    -RepoName "skyvern" `
    -PatternName "task-orchestration-browser" `
    -Description "Task orchestration for complex browser automation workflows" `
    -Category "browser" `
    -Family "orchestration" `
    -Tags "orchestration,workflow,automation"
```

---

### Week 8: Integration Project

**Mục tiêu**: Build integration project combining learnings

```
Project Idea: AI-Powered Anti-Detect Browser Agent

1. Use Donut Browser as base (Rust + Tauri)
2. Integrate browser-use patterns (LLM control)
3. Add fingerprint evasion from undetectable-fingerprint-browser
4. Add enterprise features from Skyvern
5. Expose via MCP for AI agent integration

Deliverable:
- Rust + Tauri desktop app
- LLM-powered browser control
- Fingerprint evasion
- MCP server integration
```

---

## 🧠 Key Patterns to Learn

### 1. Fingerprint Spoofing

- User agent spoofing
- Screen resolution spoofing
- Canvas fingerprinting evasion
- WebGL fingerprinting evasion
- Audio fingerprinting evasion

### 2. Anti-Detect Architecture

- Profile isolation
- Proxy rotation
- Cookie management
- Extension management
- Fingerprint randomization

### 3. LLM Browser Control

- Natural language to action
- Task planning
- Error recovery
- Context management
- Multi-step automation

### 4. Enterprise Features

- Multi-user support
- Task orchestration
- Integration patterns
- Logging and monitoring
- Security and compliance

---

## 📝 Learning Capture Template

Sau khi học mỗi pattern, capture learning vào TiBrain:

```powershell
.\add-learning-to-tibrain.ps1 `
    -RepoName "[repo-name]" `
    -PatternName "[pattern-name]" `
    -Description "[detailed description]" `
    -Category "browser" `
    -Family "[family]" `
    -Tags "[relevant tags]"
```

---

## 🔍 MCP Integration

Sau khi learnings được captured, agents có thể:

```bash
# Search browser automation learnings
tibrain_search_skills --query "fingerprint spoofing"

# List all browser learnings
curl http://localhost:1810/v1/tibrain/tool/list?category=browser

# Get specific learning
curl http://localhost:1810/v1/tibrain/tool?id=learning-donut-browser-fingerprint-spoofing
```

---

## 🎯 Success Criteria

Sau khi hoàn thành learning path, bạn sẽ có thể:

- ✅ Build anti-detect browser from scratch
- ✅ Implement fingerprint spoofing techniques
- ✅ Integrate LLM for browser control
- ✅ Build enterprise browser automation platform
- ✅ Expose browser automation via MCP
- ✅ 20+ learnings captured in TiBrain
- ✅ Working integration project

---

## 📚 Resources

### Repositories

- **Donut Browser**: `Z:\10_WORKPLACE\Ti\apps\donutbrowser\`
- **undetectable-fingerprint-browser**: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\07_Repositories\undetectable-fingerprint-browser\`
- **browser-use**: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\07_Repositories\browser-use\`
- **Skyvern**: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\07_Repositories\skyvern\`

### Documentation

- **Donut Browser**: `apps/donutbrowser/README.md`
- **Learning to TiBrain**: `Ti-learning-lab/01_Learning/LEARNING_TO_TIBRAIN.md`
- **TiBrain Hub**: `apps/router/tibrain/main.go`

### Tools

- **add-learning-to-tibrain.ps1**: `Ti-learning-lab/01_Learning/add-learning-to-tibrain.ps1`
- **TiBrain Hub**: `apps/router/tibrain/tibrain-server.exe`

---

*Last Updated: 2026-05-04*
