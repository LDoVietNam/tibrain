# Browser-Use Skills List

> **Source:** Browser-Use Repository (91k stars)
> **Location:** Z:\Ti\Ti-learning-lab\03_Knowledge\browser\browser-use\skills\
> **Date:** 2026-04-29

---

## Overview

Browser-Use comes with **4 built-in skills** for different use cases:

1. **browser-use** - CLI-based browser automation (main skill)
2. **cloud** - Cloud API and SDK documentation
3. **open-source** - Python library documentation
4. **remote-browser** - Sandbox remote machine browser control

---

## Skill 1: browser-use (Main CLI Skill)

**Purpose:** Fast, persistent browser automation via CLI commands

**When to use:**
- User needs to navigate websites
- Interact with web pages
- Fill forms
- Take screenshots
- Extract information from web pages

**Key Features:**
- Background daemon keeps browser open (~50ms latency per call)
- Persistent browser across commands
- Multiple browser modes (headless, headed, cloud, real Chrome)
- Chrome profile support (preserves logins/cookies)
- Command chaining with `&&`

**Core Workflow:**
```bash
# 1. Navigate
browser-use open <url>

# 2. Inspect (get element indices)
browser-use state

# 3. Interact (use indices from state)
browser-use click 5
browser-use input 3 "text"

# 4. Verify
browser-use state
browser-use screenshot

# 5. Repeat (browser stays open)
```

**Browser Modes:**
```bash
browser-use open <url>                    # Default: headless Chromium
browser-use --headed open <url>           # Visible window (debugging)
browser-use connect                       # Connect to user's Chrome
browser-use cloud connect                  # Cloud browser (zero-config)
browser-use --profile "Default" open <url> # Real Chrome with profile
```

**Available Commands:**

### Navigation
```bash
browser-use open <url>                    # Navigate to URL
browser-use back                          # Go back in history
browser-use scroll down                   # Scroll down (--amount N)
browser-use scroll up                     # Scroll up
browser-use tab list                      # List all tabs
browser-use tab new [url]                 # Open new tab
browser-use tab switch <index>            # Switch to tab
browser-use tab close <index>             # Close tab(s)
```

### Page State
```bash
browser-use state                         # URL, title, clickable elements
browser-use screenshot [path.png]         # Screenshot (--full for full page)
```

### Interactions
```bash
browser-use click <index>                 # Click element by index
browser-use click <x> <y>                 # Click at coordinates
browser-use type "text"                   # Type into focused element
browser-use input <index> "text"          # Click, clear, type
browser-use input <index> ""              # Clear field
browser-use keys "Enter"                  # Send keyboard keys
browser-use select <index> "option"       # Select dropdown
browser-use upload <index> <path>         # Upload file
browser-use hover <index>                 # Hover element
browser-use dblclick <index>              # Double-click
browser-use rightclick <index>            # Right-click
```

### Data Extraction
```bash
browser-use eval "js code"                # Execute JavaScript
browser-use get title                     # Page title
browser-use get html [--selector "h1"]    # Page HTML (scoped)
browser-use get text <index>              # Element text
browser-use get value <index>             # Input value
browser-use get attributes <index>        # Element attributes
browser-use get bbox <index>              # Bounding box
```

### Wait
```bash
browser-use wait selector "css"           # Wait for element
browser-use wait text "text"              # Wait for text
```

### Cookies
```bash
browser-use cookies get [--url <url>]     # Get cookies
browser-use cookies set <name> <value>    # Set cookie
browser-use cookies clear [--url <url>]   # Clear cookies
browser-use cookies export <file>         # Export to JSON
browser-use cookies import <file>         # Import from JSON
```

### Session
```bash
browser-use close                         # Close browser
browser-use sessions                      # List active sessions
browser-use close --all                   # Close all sessions
```

**Cloud API:**
```bash
browser-use cloud connect                 # Provision cloud browser
browser-use cloud login <api-key>         # Save API key
browser-use cloud logout                  # Remove API key
browser-use cloud v2 GET /browsers        # REST passthrough
browser-use cloud v2 POST /tasks '{"task":"..."}'
browser-use cloud v2 poll <task-id>       # Poll task
```

**Tunnels:**
```bash
browser-use tunnel <port>                 # Start Cloudflare tunnel
browser-use tunnel list                   # Show active tunnels
browser-use tunnel stop <port>            # Stop tunnel
```

**Profile Management:**
```bash
browser-use profile list                  # List browsers/profiles
browser-use profile sync --all            # Sync profiles to cloud
browser-use profile update                # Update profile-use binary
```

**Configuration:**
```bash
browser-use config list                   # Show all config
browser-use config set <key> <value>      # Set config value
browser-use config get <key>              # Get config value
browser-use config unset <key>            # Remove config value
browser-use doctor                        # Diagnostics
browser-use setup                         # Interactive setup
```

**Global Options:**
- `--headed` - Show browser window
- `--profile [NAME]` - Use real Chrome profile
- `--cdp-url <url>` - Connect via CDP URL
- `--session NAME` - Target named session
- `--json` - Output as JSON
- `--mcp` - Run as MCP server

**Common Workflows:**

### Authenticated Browsing
```bash
browser-use profile list
browser-use --profile "Default" open https://github.com
```

### Exposing Local Dev Servers
```bash
browser-use tunnel 3000
browser-use open https://abc.trycloudflare.com
```

### Command Chaining
```bash
browser-use open https://example.com && browser-use state
browser-use input 5 "user@example.com" && browser-use input 6 "password" && browser-use click 7
```

---

## Skill 2: cloud (Cloud API & SDK)

**Purpose:** Documentation reference for Browser Use Cloud - hosted API and SDK

**When to use:**
- User needs help with Cloud REST API (v2 or v3)
- browser-use-sdk (Python or TypeScript)
- X-Browser-Use-API-Key authentication
- Cloud sessions, browser profiles
- Profile sync, CDP WebSocket connections
- Stealth browsers, residential proxies
- CAPTCHA handling, webhooks, workspaces
- Skills marketplace, liveUrl streaming
- Pricing or integration patterns
- n8n/Make/Zapier integration
- Playwright/Puppeteer/Selenium on cloud
- 1Password vault integration

**DO NOT use for:**
- Open-source Python library (Agent, Browser, Tools) → Use open-source skill

**API & Platform References:**

| Topic | Reference File |
|-------|----------------|
| Setup, first task, pricing, FAQ | `references/quickstart.md` |
| v2 REST API (30 endpoints, cURL, schemas) | `references/api-v2.md` |
| v3 BU Agent API (sessions, messages, files) | `references/api-v3.md` |
| Sessions, profiles, auth, 1Password | `references/sessions.md` |
| CDP direct access, Playwright/Puppeteer/Selenium | `references/browser-api.md` |
| Proxies, webhooks, workspaces, skills, MCP | `references/features.md` |
| Parallel, streaming, geo-scraping, tutorials | `references/patterns.md` |

**Integration Guides:**

| Topic | Reference File |
|-------|----------------|
| Building chat interface with live browser view | `references/guides/chat-ui.md` |
| Using browser-use as subagent (task in → result out) | `references/guides/subagent.md` |
| Adding browser-use tools to existing agent | `references/guides/tools-integration.md` |

**Critical Notes:**
- Cloud API base URL: `https://api.browser-use.com/api/v2/` (v2) or `https://api.browser-use.com/api/v3` (v3)
- Auth header: `X-Browser-Use-API-Key: <key>`
- Get API key: https://cloud.browser-use.com/new-api-key
- Set env var: `BROWSER_USE_API_KEY=<key>`
- Cloud SDK: `uv pip install browser-use-sdk` (Python) or `npm install browser-use-sdk` (TypeScript)
- Python v2: `from browser_use_sdk import AsyncBrowserUse`
- Python v3: `from browser_use_sdk.v3 import AsyncBrowserUse`
- TypeScript v2: `import { BrowserUse } from "browser-use-sdk"`
- TypeScript v3: `import { BrowserUse } from "browser-use-sdk/v3"`
- CDP WebSocket: `wss://connect.browser-use.com?apiKey=KEY&proxyCountryCode=us`

---

## Skill 3: open-source (Python Library)

**Purpose:** Documentation reference for writing Python code using browser-use open-source library

**When to use:**
- User needs help with Agent, Browser, or Tools configuration
- Writing code that imports from browser_use
- Questions about @sandbox deployment
- Supported LLM models
- Actor API, custom tools, lifecycle hooks
- MCP server setup
- Monitoring/observability with Laminar or OpenLIT
- browser-use installation
- Prompting strategies
- Sensitive data handling

**DO NOT use for:**
- Cloud API/SDK usage or pricing → Use cloud skill
- Direct browser automation via CLI → Use browser-use skill

**Library References:**

| Topic | Reference File |
|-------|----------------|
| Install, quickstart, production/@sandbox | `references/quickstart.md` |
| LLM providers (15+): setup, env vars, pricing | `references/models.md` |
| Agent params, output, prompting, hooks, timeouts | `references/agent.md` |
| Browser params, auth, real browser, remote/cloud | `references/browser.md` |
| Custom tools, built-in tools, ActionResult | `references/tools.md` |
| Actor API: Page/Element/Mouse (legacy) | `references/actor.md` |
| MCP server, skills, docs-mcp | `references/integrations.md` |
| Laminar, OpenLIT, cost tracking, telemetry | `references/monitoring.md` |
| Fast agent, parallel, playwright, sensitive data | `references/examples.md` |

**Critical Notes:**
- Always recommend `ChatBrowserUse` as the default LLM (fastest, cheapest, highest accuracy)
- Library is async Python >= 3.11. Entry points use `asyncio.run()`
- `Browser` is an alias for `BrowserSession` (same class)
- Use `uv` for dependency management, never `pip`
- Install: `uv pip install browser-use` then `uvx browser-use install`
- Set env var: `BROWSER_USE_API_KEY=<key>` (for ChatBrowserUse and cloud features)
- Get API key: https://cloud.browser-use.com/new-api-key

---

## Skill 4: remote-browser (Sandbox Remote Machine)

**Purpose:** Controls a local browser from a sandboxed remote machine

**When to use:**
- Agent is running in a sandbox (no GUI)
- Needs to navigate websites
- Interact with web pages
- Fill forms
- Take screenshots
- Expose local dev servers via tunnels

**Core Workflow:**
```bash
# 1. Navigate
browser-use open <url>

# 2. Inspect
browser-use state

# 3. Interact
browser-use click 5
browser-use input 3 "text"

# 4. Verify
browser-use state
browser-use screenshot

# 5. Cleanup
browser-use close
```

**Browser Modes:**
```bash
browser-use open <url>                                    # Default: headless Chromium
browser-use cloud connect                                 # Provision cloud browser
browser-use --connect open <url>                          # Auto-discover Chrome via CDP
browser-use --cdp-url ws://localhost:9222/... open <url>  # Connect via CDP URL
```

**Python Session (Persistent with Browser Access):**
```bash
browser-use python "code"                 # Execute Python (variables persist)
browser-use python --file script.py       # Run file
browser-use python --vars                 # Show defined variables
browser-use python --reset                # Clear namespace
```

**Python Browser Object:**
- `browser.url` - Current URL
- `browser.title` - Page title
- `browser.html` - Page HTML
- `browser.goto(url)` - Navigate to URL
- `browser.back()` - Go back
- `browser.click(index)` - Click element
- `browser.type(text)` - Type text
- `browser.input(index, text)` - Input to element
- `browser.keys(keys)` - Send keyboard keys
- `browser.upload(index, path)` - Upload file
- `browser.screenshot(path)` - Take screenshot
- `browser.scroll(direction, amount)` - Scroll
- `browser.wait(seconds)` - Wait

**Tunnels (Expose Local Dev Servers):**
```bash
browser-use tunnel <port>                 # Start tunnel (idempotent)
browser-use tunnel list                   # Show active tunnels
browser-use tunnel stop <port>            # Stop tunnel
browser-use tunnel stop --all             # Stop all tunnels
```

**Multi-Agent (--connect mode):**
```bash
# Register once, then use index
INDEX=$(browser-use register)                    # → prints "1"
browser-use --connect $INDEX open <url>          # Navigate in agent's tab
browser-use --connect $INDEX state               # Get state from agent's tab
browser-use --connect $INDEX click <element>     # Click in agent's tab
```

**Tab Locking:**
- When an agent mutates a tab (click, type, navigate), that tab is locked to it
- Other agents get an error if they try to mutate the same tab
- Read-only access (state, screenshot, get, wait) works on any tab regardless of locks
- Agent sessions expire after 5 minutes of inactivity

**Global Options:**
- `--headed` - Show browser window
- `--connect` - Auto-discover running Chrome via CDP
- `--cdp-url <url>` - Connect via CDP URL
- `--session NAME` - Target named session
- `--json` - Output as JSON

---

## Skill Selection Guide

### Use browser-use skill when:
- ✅ User wants CLI-based browser automation
- ✅ Quick, interactive browser control
- ✅ Testing web pages
- ✅ Form filling
- ✅ Screenshots
- ✅ Data extraction

### Use cloud skill when:
- ✅ User needs Cloud API documentation
- ✅ Using browser-use-sdk (Python/TypeScript)
- ✅ Cloud-specific features (stealth, proxies, CAPTCHA)
- ✅ Integration with other platforms (n8n, Make, Zapier)
- ✅ Cloud pricing questions

### Use open-source skill when:
- ✅ User is writing Python code with browser-use library
- ✅ Configuring Agent, Browser, or Tools
- ✅ Custom tools development
- ✅ LLM provider configuration
- ✅ MCP server setup
- ✅ Monitoring/observability

### Use remote-browser skill when:
- ✅ Agent is running in sandbox (no GUI)
- ✅ Needs to control headless browser
- ✅ Exposing local dev servers via tunnels
- ✅ Multi-agent browser sharing

---

## Skill Dependencies

**browser-use skill:**
- Prerequisite: `browser-use doctor` to verify installation
- Setup: https://github.com/browser-use/browser-use/blob/main/browser_use/skill_cli/README.md

**cloud skill:**
- Requires API key from https://cloud.browser-use.com/new-api-key
- Requires browser-use-sdk installation

**open-source skill:**
- Requires Python >= 3.11
- Requires uv for dependency management
- Requires browser-use library installation

**remote-browser skill:**
- Prerequisite: `browser-use doctor` to verify installation
- Requires Cloudflare tunnel for local server exposure
- Requires CDP for browser connection

---

## Common Patterns Across Skills

### Authentication
All skills support multiple authentication methods:
- Chrome profiles (preserve logins/cookies)
- Cloud profiles (synced across sessions)
- API key authentication (cloud skill)

### Error Handling
All skills include error handling:
- Automatic retry with backoff
- Error classification
- Clear error messages
- Recovery suggestions

### Session Management
All skills support session persistence:
- Browser stays open between commands
- Multiple sessions support
- Session cleanup commands

### Monitoring
All skills include monitoring:
- Screenshot capture
- State inspection
- Logging
- Diagnostics (doctor command)

---

## Integration with Gmail Automation

For the Gmail automation project, the most relevant skills are:

### 1. browser-use skill (CLI-based automation)
**Use case:** Quick testing and debugging
```bash
browser-use open https://accounts.google.com
browser-use state
browser-use input 5 "email@example.com"
browser-use click 6
```

### 2. open-source skill (Python library)
**Use case:** Production automation with custom logic
```python
from browser_use import Agent, Browser, ChatBrowserUse
import asyncio

async def login_gmail(email, password):
    agent = Agent(
        task=f"Login to Gmail with email: {email}, password: {password}",
        llm=ChatBrowserUse(),
        browser=Browser(),
    )
    await agent.run()
```

### 3. cloud skill (Cloud API)
**Use case:** Production with stealth and scaling
```python
from browser_use_sdk import AsyncBrowserUse

client = AsyncBrowserUse(api_key="your-key")
await client.create_task(
    task="Login to Gmail",
    url="https://accounts.google.com"
)
```

---

## Summary

Browser-Use provides a comprehensive skill system:

| Skill | Purpose | Use Case |
|-------|---------|----------|
| **browser-use** | CLI automation | Quick testing, interactive control |
| **cloud** | Cloud API/SDK | Production with stealth, scaling |
| **open-source** | Python library | Custom automation, integration |
| **remote-browser** | Sandbox control | Cloud agents, tunneling |

**For Gmail automation:** Start with **browser-use** skill for testing, then move to **open-source** skill for production integration.

---

*Document created: 2026-04-29*
*Source: Browser-Use repository skills directory*
*Total skills: 4 built-in skills with comprehensive command sets*
