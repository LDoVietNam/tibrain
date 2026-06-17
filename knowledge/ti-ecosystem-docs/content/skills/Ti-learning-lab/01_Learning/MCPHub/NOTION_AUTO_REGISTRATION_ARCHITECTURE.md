# Notion Auto-Registration POC - Integration Architecture

## Overview

This document describes the integration architecture for a Notion account auto-registration POC that combines learnings from 4 browser automation repositories:

1. **Donut Browser** (Rust + Tauri) - Core browser management with fingerprint spoofing
2. **donutbrowser-go** - Go client for REST API integration
3. **undetectable-fingerprint-browser** - Advanced fingerprint evasion techniques
4. **browser-use** - LLM-driven browser automation
5. **Skyvern** - Vision LLM browser automation for complex workflows

## Architecture Goals

- **Stealth**: Comprehensive fingerprint spoofing to avoid detection
- **Automation**: LLM-driven automation for complex multi-step workflows
- **Scalability**: Multi-account registration with profile isolation
- **Reliability**: Error handling, retry logic, and state persistence
- **Flexibility**: Support for different automation approaches (LLM vs traditional)

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Ti CLI (Go Microkernel)                       │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Notion Automation Module (Go)                           │  │
│  │  - donutbrowser-go client integration                    │  │
│  │  - Profile management                                   │  │
│  │  - Proxy rotation                                        │  │
│  │  - Task orchestration                                    │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ REST API
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│              Donut Browser (Rust + Tauri)                       │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Profile Manager                                         │  │
│  │  - UUID-based isolation                                  │  │
│  │  - Camoufox/Wayfern fingerprint configs                  │  │
│  │  - Proxy/VPN per profile                                 │  │
│  │  - DNS blocklist integration                             │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  MCP Server                                              │  │
│  │  - JSON-RPC 2.0 API                                      │  │
│  │  - Agent control interface                               │  │
│  │  - Tool discovery                                        │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ CDP / WebSocket
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│              Browser Engine (Chromium/Firefox)                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Fingerprint Spoofing Layer                              │  │
│  │  - Canvas, WebGL, AudioContext spoofing                  │  │
│  │  - Font, Timezone, Hardware spoofing                    │  │
│  │  - WebRTC leak prevention                                │  │
│  │  - GPS/Sensor data emulation                            │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ AI Control
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│              Automation Engine (Python)                          │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  browser-use Integration                                │  │
│  │  - LLM-driven action planning                            │  │
│  │  - CDP integration via cdp-use                           │  │
│  │  - Event-driven browser management                       │  │
│  │  - Multi-LLM provider support                            │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Skyvern Integration (Optional)                          │  │
│  │  - Vision LLM for element detection                      │  │
│  │  - Playwright AI extension                               │  │
│  │  - Workflow engine for complex tasks                      │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ Target Site
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Notion.com                                    │
│  - Microsoft SSO authentication                                │
│  - Account creation workflow                                    │
│  - CAPTCHA challenges                                           │
└─────────────────────────────────────────────────────────────────┘
```

## Component Integration

### 1. Ti CLI - Orchestration Layer

**Purpose**: Main orchestrator for automation tasks

**Key Components**:
- `internal/donutbrowser/client.go` - REST API client
- `cmd/notion.go` - Notion-specific automation commands
- Profile management via Go client
- Task queue and retry logic

**Integration Points**:
- Donut Browser REST API (port 10108)
- MCP server for AI agent control
- Proxy rotation service

### 2. Donut Browser - Browser Management

**Purpose**: Browser lifecycle and fingerprint management

**Key Components**:
- `ProfileManager` - Profile creation, isolation, sync
- `CamoufoxManager` - Firefox fingerprint spoofing
- `WayfernManager` - Chromium fingerprint spoofing
- `ProxyManager` - Per-profile proxy configuration
- `McpServer` - AI agent control interface

**Integration Points**:
- Go client via REST API
- MCP server for AI agents
- Browser engines (Camoufox, Wayfern)

### 3. undetectable-fingerprint-browser - Fingerprint Library

**Purpose**: Advanced fingerprint spoofing techniques

**Integration Approach**:
- Extract fingerprint data structures (user-agents.json, webgl.json)
- Integrate consistency analysis engine
- Apply anti-leak modules (WebRTC, Canvas/WebGL spoofing)
- Use as reference for Donut Browser config generation

### 4. browser-use - LLM Automation Engine

**Purpose**: AI-driven browser automation

**Key Components**:
- `Agent` - Main orchestrator with LLM decision loop
- `BrowserSession` - CDP connection management
- `Tools` - Action registry (click, type, scroll, etc.)
- `DomService` - DOM content extraction

**Integration Points**:
- Connect to Donut Browser via CDP
- Use MCP server for tool discovery
- Multi-LLM provider support (ChatBrowserUse recommended)
- Event-driven browser state management

### 5. Skyvern - Vision LLM Automation (Optional)

**Purpose**: Advanced vision-based automation for complex workflows

**Key Components**:
- `Agent System` - Vision LLM-based agent loop
- `Playwright AI Extension` - Natural language actions
- `Workflow Engine` - No-code workflow builder

**Integration Points**:
- Use as alternative to browser-use for complex tasks
- Leverage Playwright AI extension for natural language actions
- Cloud platform for anti-bot detection and CAPTCHA solving

## Workflow: Notion Account Auto-Registration

### Phase 1: Profile Creation

1. **Generate Fingerprint**
   - Use undetectable-fingerprint-browser data as reference
   - Generate consistent fingerprint (Canvas, WebGL, Audio, etc.)
   - Select realistic user agent and device profile
   - Apply anti-leak configurations

2. **Create Donut Browser Profile**
   - Call Go client: `client.CreateProfile()`
   - Configure Camoufox or Wayfern with fingerprint
   - Assign proxy (rotate per profile)
   - Set DNS blocklist level
   - Add tags: "notion-automation", "auto-registration"

3. **Launch Profile**
   - Call Go client: `client.RunProfile(profileID)`
   - Wait for browser to be ready
   - Get CDP port for automation connection

### Phase 2: LLM Automation

**Option A: browser-use Integration**

```python
from browser_use import Agent, ChatBrowserUse
from playwright.sync_api import sync_playwright

# Connect to Donut Browser via CDP
with sync_playwright() as p:
    browser = p.chromium.connect_over_cdp(f"http://localhost:{cdp_port}")
    page = browser.new_page()
    
    # Initialize browser-use agent
    agent = Agent(
        task="Register a Notion account with the following email: {email}",
        llm=ChatBrowserUse(),
        browser_session=page,
    )
    
    # Run automation
    history = await agent.run(max_steps=50)
    
    # Extract credentials
    credentials = extract_credentials(history)
```

**Option B: Skyvern Integration**

```python
from skyvern import Skyvern

skyvern = Skyvern(
    api_key="your-key",
    workflow_id="notion-registration-workflow"
)

# Execute pre-built workflow
result = skyvern.agent.run_task(
    prompt=f"Register Notion account with email {email}",
    workflow_id="notion-registration-workflow"
)

# Extract credentials
credentials = result.extracted_data
```

### Phase 3: Credential Extraction

1. **Cookie Extraction**
   - Use Donut Browser cookie manager
   - Extract authentication cookies
   - Store securely in TiBrain

2. **Session Persistence**
   - Save profile state for reuse
   - Enable sync mode (encrypted)
   - Store in TiBrain memory

### Phase 4: Cleanup

1. **Kill Profile**
   - Call Go client: `client.KillProfile(profileID)`
   - Clean up browser processes
   - Release CDP port

2. **Archive Profile**
   - Mark profile as "registered"
   - Store metadata in TiBrain
   - Enable for future use

## Data Flow

```
User Request (Email)
    ↓
Ti CLI (Go)
    ↓
Donut Browser API (Create Profile)
    ↓
Profile with Fingerprint + Proxy
    ↓
Launch Browser (CDP Port)
    ↓
browser-use Agent (Python)
    ↓
LLM Decision Loop
    ↓
Browser Actions (via CDP)
    ↓
Notion Registration Complete
    ↓
Credential Extraction
    ↓
Store in TiBrain
    ↓
Cleanup Profile
```

## Error Handling & Retry Logic

### 1. Fingerprint Generation
- **Error**: Invalid fingerprint configuration
- **Retry**: Regenerate with different profile
- **Fallback**: Use default fingerprint

### 2. Profile Creation
- **Error**: Proxy connection failed
- **Retry**: Rotate to next proxy
- **Fallback**: Create without proxy

### 3. Browser Launch
- **Error**: Browser failed to start
- **Retry**: Restart profile
- **Fallback**: Use different browser engine

### 4. LLM Automation
- **Error**: LLM rate limit
- **Retry**: Switch to secondary LLM
- **Fallback**: Use traditional automation

### 5. CAPTCHA
- **Error**: CAPTCHA detection
- **Retry**: Use Skyvern Cloud CAPTCHA solver
- **Fallback**: Manual intervention

## Security Considerations

1. **Credential Storage**
   - Encrypt stored credentials
   - Use TiBrain secure storage
   - Never log sensitive data

2. **Proxy Security**
   - Use reputable proxy providers
   - Rotate proxies regularly
   - Monitor for proxy leaks

3. **Fingerprint Consistency**
   - Validate fingerprint before use
   - Use consistency analysis engine
   - Avoid detection signals

4. **Rate Limiting**
   - Respect Notion rate limits
   - Implement exponential backoff
   - Monitor for account flags

## Performance Optimization

1. **Parallel Registration**
   - Launch multiple profiles concurrently
   - Use different proxies per profile
   - Limit concurrency to avoid detection

2. **Profile Reuse**
   - Reuse successful profiles
   - Maintain fingerprint consistency
   - Update profiles periodically

3. **LLM Optimization**
   - Use ChatBrowserUse for speed
   - Cache LLM responses
   - Use secondary LLM for lightweight tasks

4. **Browser Optimization**
   - Use headless mode when possible
   - Disable unnecessary features
   - Optimize CDP connection

## Monitoring & Observability

1. **Metrics to Track**
   - Registration success rate
   - Time per registration
   - LLM token usage
   - Proxy success rate
   - Fingerprint detection rate

2. **Logging**
   - Profile creation logs
   - LLM decision logs
   - Error logs with context
   - Performance metrics

3. **Alerts**
   - High failure rate
   - Proxy exhaustion
   - LLM rate limits
   - Detection events

## Next Steps

1. **Implement Ti CLI Module**
   - Extend `internal/donutbrowser/client.go`
   - Add notion-specific commands
   - Implement task orchestration

2. **Integrate browser-use**
   - Set up Python environment
   - Configure LLM provider
   - Test CDP connection to Donut Browser

3. **Develop Workflow**
   - Create Notion registration workflow
   - Test with single account
   - Scale to multiple accounts

4. **Add Monitoring**
   - Implement metrics collection
   - Set up logging
   - Configure alerts

5. **Production Readiness**
   - Security audit
   - Performance testing
   - Documentation

## References

- Donut Browser: `Z:\10_WORKPLACE\Ti\apps\donutbrowser\`
- donutbrowser-go: `Z:\10_WORKPLACE\Ti\apps\cli\internal\donutbrowser\`
- undetectable-fingerprint-browser: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\07_Repositories\undetectable-fingerprint-browser\`
- browser-use: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\07_Repositories\browser-use\`
- Skyvern: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\07_Repositories\skyvern\`
