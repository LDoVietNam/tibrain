# Browser Automation Lessons Learned

> **Source Repositories:**
> - Browser-Use (91k stars) - Z:\Ti\Ti-learning-lab\03_Knowledge\browser\browser-use
> - Skyvern (21.4k stars) - Z:\Ti\Ti-learning-lab\03_Knowledge\browser\skyvern
>
> **Date:** 2026-04-29
> **Purpose:** Extract actionable patterns and lessons for browser automation projects

---

## Executive Summary

Both repositories represent state-of-the-art approaches to AI-powered browser automation, but with fundamentally different philosophies:

**Browser-Use** follows a **simple, agent-focused** approach:
- Natural language task definition
- LLM-driven decision making
- Minimal overhead
- Perfect for repetitive, well-defined tasks

**Skyvern** follows an **enterprise-grade, workflow-focused** approach:
- Computer vision for element detection
- Workflow orchestration engine
- Database persistence
- Full-stack architecture (Python backend + Node.js frontend)
- Designed for complex, multi-step workflows

**Key Insight:** The right tool depends on task complexity. Simple tasks benefit from Browser-Use's simplicity; complex enterprise workflows benefit from Skyvern's orchestration.

---

## Browser-Use Lessons

### 1. Architecture Patterns

#### Event-Driven Browser Session
**Pattern:** Browser sessions are event-driven with backwards compatibility.

```python
# Key components:
# - EventBus for async communication
# - CDPClient for Chrome DevTools Protocol
# - Event types: BrowserLaunchEvent, NavigationCompleteEvent, etc.
from bubus import EventBus
from cdp_use import CDPClient
```

**Lesson:** Event-driven architecture enables:
- Loose coupling between components
- Easy extensibility (add new event handlers)
- Better observability (track all browser events)
- Backwards compatibility (old APIs still work)

**Applicability:** Use event-driven architecture for:
- Browser automation
- Any system with multiple independent components
- Systems requiring extensibility

#### Agent-First Design
**Pattern:** Everything revolves around the Agent class.

```python
agent = Agent(
    task="Find the number of stars of the browser-use repo",
    llm=ChatBrowserUse(),
    browser=browser,
)
await agent.run()
```

**Lesson:** Agent-first design means:
- Simple API: one task string → execution
- LLM handles all decision making
- No need to define step-by-step workflows
- Natural language interface

**Applicability:** Use agent-first when:
- Tasks are well-defined but execution varies
- You want LLM to handle decision logic
- You prefer natural language over code

#### LLM Abstraction Layer
**Pattern:** Multiple LLM providers through unified interface.

```python
from browser_use import ChatBrowserUse, ChatGoogle, ChatAnthropic

# All implement same interface
llm = ChatBrowserUse()  # Optimized for browser automation
llm = ChatGoogle(model='gemini-3-flash-preview')
llm = ChatAnthropic(model='claude-sonnet-4-6')
```

**Lesson:** LLM abstraction enables:
- Easy provider switching
- Cost optimization (use cheaper models for simple tasks)
- Model routing based on task complexity
- Testing with different models

**Applicability:** Always abstract LLM providers when:
- Building LLM-powered applications
- You might switch providers
- You want cost optimization

---

### 2. Browser Control Patterns

#### CDP (Chrome DevTools Protocol) Integration
**Pattern:** Direct CDP integration for fine-grained browser control.

```python
from cdp_use import CDPClient
from cdp_use.cdp.target.commands import CreateTargetParameters
```

**Lesson:** CDP provides:
- Full browser control (tabs, navigation, cookies)
- Performance monitoring
- Network interception
- JavaScript execution

**Applicability:** Use CDP when:
- You need fine-grained browser control
- You want to monitor network traffic
- You need to modify page behavior

#### Watchdog Pattern
**Pattern:** Multiple watchdogs monitor browser state and react to issues.

```python
# Watchdogs in browser-use:
# - captcha_watchdog.py - Detect CAPTCHAs
# - crash_watchdog.py - Detect browser crashes
# - dom_watchdog.py - Detect DOM changes
# - popup_watchdog.py - Handle popups
# - security_watchdog.py - Detect security issues
```

**Lesson:** Watchdog pattern enables:
- Proactive issue detection
- Automatic error recovery
- Separation of concerns (each watchdog handles one issue)
- Easy to add new monitors

**Applicability:** Use watchdogs when:
- System has multiple failure modes
- You want automatic recovery
- You need proactive monitoring

#### Profile Management
**Pattern:** Browser profiles for session persistence.

```python
from browser_use import BrowserProfile

profile = BrowserProfile(
    headless=True,
    proxy_settings=ProxySettings(...)
)
browser = Browser(profile=profile)
```

**Lesson:** Profile management provides:
- Session persistence (cookies, local storage)
- Proxy configuration
- Headless vs headed mode
- Custom browser configurations

**Applicability:** Use profiles when:
- You need session persistence
- You use proxies
- You need different browser configurations

---

### 3. DOM Analysis Patterns

#### Enhanced DOM Snapshot
**Pattern:** Rich DOM representation with visual information.

```python
from browser_use.dom.service import DomService

# Provides:
# - Element hierarchy
# - Visual attributes (position, size, visibility)
# - Interactive elements
# - Text content
```

**Lesson:** Enhanced DOM enables:
- Better element identification
- Visual understanding of page
- Robust element selection (not just XPath)
- LLM-friendly representation

**Applicability:** Use enhanced DOM when:
- You need robust element selection
- Pages have dynamic content
- You want LLM to understand page structure

#### Clickable Elements Detection
**Pattern:** Automatic detection of interactive elements.

```python
from browser_use.dom.serializer.clickable_elements import ClickableElements

# Automatically detects:
# - Buttons
# - Links
# - Inputs
# - Clickable divs
```

**Lesson:** Clickable detection enables:
- No manual selector definition
- Robust to layout changes
- Better UX for LLM agents

**Applicability:** Use when:
- You want robust element selection
- You don't want to maintain selectors
- Pages change frequently

---

### 4. Error Handling Patterns

#### Agent Error Classification
**Pattern:** Structured error types for different failure modes.

```python
from browser_use.agent.views import AgentError

# Error types include:
# - Navigation errors
# - Element not found
# - Timeout errors
# - LLM errors
```

**Lesson:** Error classification enables:
- Specific error handling per type
- Better error messages
- Targeted retry logic
- Improved debugging

**Applicability:** Always classify errors when:
- System has multiple failure modes
- You need specific handling per error type
- You want good error messages

#### Retry with Exponential Backoff
**Pattern:** Automatic retry with increasing delays.

```python
# Built into agent execution
# Retries on:
# - Network errors
# - LLM rate limits
# - Temporary failures
```

**Lesson:** Retry with backoff provides:
- Resilience to transient failures
- Respect for rate limits
- Better success rates
- No manual retry logic

**Applicability:** Always use retry with backoff for:
- Network operations
- LLM API calls
- Any external service calls

---

### 5. Performance Patterns

#### Token Cost Tracking
**Pattern:** Track LLM token usage and costs.

```python
from browser_use.tokens.service import TokenCost

# Tracks:
# - Input tokens
# - Output tokens
# - Cost per provider
# - Total cost
```

**Lesson:** Cost tracking enables:
- Budget management
- Cost optimization
- Provider comparison
- Usage analytics

**Applicability:** Always track costs when:
- Using LLM APIs
- You have budget constraints
- You want to optimize costs

#### Message Compaction
**Pattern:** Compact message history to reduce token usage.

```python
from browser_use.agent.views import MessageCompactionSettings

# Compacts:
# - Removes redundant messages
# - Summarizes long conversations
# - Keeps context window manageable
```

**Lesson:** Message compaction provides:
- Lower token costs
- Faster LLM responses
- Longer conversations possible
- Better performance

**Applicability:** Use compaction when:
- Conversations are long
- Token costs matter
- You want to maintain context

---

### 6. Testing Patterns

#### Demo Mode
**Pattern:** Visual demo mode for testing and debugging.

```python
from browser_use.browser.demo_mode import DemoMode

# Provides:
# - Visual browser (not headless)
# - Slow execution (see what's happening)
# - Screenshot on each step
```

**Lesson:** Demo mode enables:
- Easy debugging
- Visual verification
- Better understanding of agent behavior
- User-friendly testing

**Applicability:** Use demo mode when:
- Debugging agent behavior
- Demonstrating to users
- Testing new features

#### Screenshot Capture
**Pattern:** Automatic screenshots for debugging.

```python
# Built into agent execution
# Captures screenshots on:
# - Each step
# - Errors
# - Completion
```

**Lesson:** Screenshots provide:
- Visual debugging
- Evidence of execution
- Better error understanding
- User feedback

**Applicability:** Always capture screenshots when:
- Debugging browser automation
- You need visual verification
- You want to show execution to users

---

### 7. Integration Patterns

#### Custom Tools
**Pattern:** Extend agent capabilities with custom tools.

```python
from browser_use import Tools

tools = Tools()

@tools.action(description='Description of what this tool does.')
def custom_tool(param: str) -> str:
    return f"Result: {param}"

agent = Agent(
    task="Your task",
    llm=llm,
    browser=browser,
    tools=tools,
)
```

**Lesson:** Custom tools enable:
- Domain-specific capabilities
- Integration with external services
- Custom actions beyond browser control
- Extensibility

**Applicability:** Use custom tools when:
- You need domain-specific actions
- You want to integrate external services
- Browser actions aren't enough

#### Cloud Integration
**Pattern:** Seamless integration with cloud services.

```python
browser = Browser(
    use_cloud=True,  # Use stealth browser on cloud
)
```

**Lesson:** Cloud integration provides:
- Stealth browsers (avoid detection)
- Proxy rotation
- CAPTCHA solving
- Scalability

**Applicability:** Use cloud when:
- You need stealth
- You use proxies
- You want to avoid CAPTCHAs
- You need scalability

---

## Skyvern Lessons

### 1. Architecture Patterns

#### Agent Swarm Pattern
**Pattern:** Multiple specialized agents working together.

```
Skyvern uses a swarm of agents to:
- Comprehend website structure
- Plan actions
- Execute actions
- Validate results
```

**Lesson:** Agent swarm enables:
- Specialization (each agent is good at one thing)
- Parallel execution
- Better error handling
- More complex workflows

**Applicability:** Use agent swarm when:
- Tasks are complex
- You need parallel execution
- You want specialization

#### Full-Stack Architecture
**Pattern:** Python backend + Node.js frontend.

```
Backend (Python):
- Agent orchestration
- Browser control (Playwright)
- Workflow engine
- API layer

Frontend (Node.js):
- Workflow builder UI
- Dashboard
- Real-time monitoring
```

**Lesson:** Full-stack architecture provides:
- Separate concerns (backend logic, frontend UI)
- Technology choice per layer (Python for AI, JS for UI)
- Team specialization (backend devs, frontend devs)
- Better user experience

**Applicability:** Use full-stack when:
- You need a UI
- You have a team
- You want technology flexibility

#### Workflow Engine
**Pattern:** Declarative workflow definition.

```python
# Workflow blocks:
# - Navigation blocks
# - Action blocks
# - Validation blocks
# - Conditional blocks
# - Loop blocks
```

**Lesson:** Workflow engine enables:
- Visual workflow building
- Non-technical users can create workflows
- Reusable workflows
- Workflow versioning

**Applicability:** Use workflow engine when:
- You need visual workflow building
- Non-technical users use the system
- You want reusable workflows

---

### 2. Computer Vision Patterns

#### Vision-Based Element Detection
**Pattern:** Use Vision LLMs to detect elements, not just DOM parsing.

```python
# Skyvern uses Vision LLMs to:
# - Identify buttons by appearance
# - Read text from images
# - Understand layout
# - Find elements without selectors
```

**Lesson:** Vision-based detection provides:
- Robustness to layout changes
- No selector maintenance
- Works on any website
- Better for complex UIs

**Applicability:** Use vision when:
- Websites change frequently
- You don't want to maintain selectors
- UIs are complex
- You need robustness

#### Screenshot Analysis
**Pattern:** Analyze screenshots to understand page state.

```python
# Skyvern captures screenshots and:
# - Passes to Vision LLM
# - Extracts information
# - Validates actions
```

**Lesson:** Screenshot analysis enables:
- Visual understanding
- CAPTCHA detection
- Layout understanding
- Better error detection

**Applicability:** Use screenshot analysis when:
- You need visual understanding
- You want to detect CAPTCHAs
- Layout matters for your task

---

### 3. Database Patterns

#### Persistent State
**Pattern:** Database persistence for all state.

```python
# Skyvern uses PostgreSQL for:
# - Workflow definitions
# - Execution history
# - User credentials
# - Agent state
```

**Lesson:** Database persistence provides:
- State recovery after crashes
- Historical analysis
- Multi-user support
- Workflow versioning

**Applicability:** Use database when:
- You need persistence
- You have multiple users
- You want historical analysis
- You need crash recovery

#### Migration System
**Pattern:** Alembic for database migrations.

```python
# Skyvern uses Alembic for:
# - Schema versioning
# - Automatic migrations
# - Rollback support
```

**Lesson:** Migration system provides:
- Safe schema changes
- Version control for schema
- Easy rollbacks
- Team collaboration

**Applicability:** Always use migrations when:
- You have a database
- Schema changes over time
- Multiple developers work on it

---

### 4. API Design Patterns

#### REST API
**Pattern:** RESTful API for all operations.

```python
# Skyvern provides REST endpoints for:
# - Workflow CRUD
# - Execution control
# - Status monitoring
# - Credential management
```

**Lesson:** REST API provides:
- Standard interface
- Easy integration
- Language agnostic
- Good tooling

**Applicability:** Use REST API when:
- You need external integration
- You want standard interface
- Multiple clients need access

#### WebSocket for Real-Time Updates
**Pattern:** WebSocket for real-time execution updates.

```python
# Skyvern uses WebSockets for:
# - Real-time execution logs
# - Live status updates
# - Interactive debugging
```

**Lesson:** WebSocket provides:
- Real-time communication
- Lower latency than polling
- Better UX for long-running tasks
- Interactive debugging

**Applicability:** Use WebSocket when:
- You need real-time updates
- Tasks are long-running
- You want interactive debugging

---

### 5. Security Patterns

#### Credential Management
**Pattern:** Secure credential storage and usage.

```python
# Skyvern supports:
# - Bitwarden integration
# - 1Password integration
# - Encrypted storage
# - Credential rotation
```

**Lesson:** Credential management provides:
- Security (no hardcoded credentials)
- Convenience (auto-fill)
- Rotation support
- Multi-provider support

**Applicability:** Always use credential management when:
- You handle sensitive data
- You need security
- You use multiple services

#### Anti-Bot Detection
**Pattern:** Built-in anti-bot detection mechanisms.

```python
# Skyvern Cloud provides:
# - Stealth browsers
# - Proxy rotation
# - CAPTCHA solving
# - Fingerprint randomization
```

**Lesson:** Anti-bot detection enables:
- Avoid detection
- Higher success rates
- Better stealth
- CAPTCHA handling

**Applicability:** Use anti-bot when:
- You scrape websites
- You need stealth
- CAPTCHAs are a problem

---

### 6. Deployment Patterns

#### Docker Compose
**Pattern:** Full containerized deployment with Docker Compose.

```yaml
# docker-compose.yml includes:
# - PostgreSQL
# - Redis
# - Backend API
# - Frontend
# - Playwright browsers
```

**Lesson:** Docker Compose provides:
- Easy local development
- Consistent environments
- One-command deployment
- Service orchestration

**Applicability:** Use Docker Compose when:
- You have multiple services
- You want easy deployment
- You need consistent environments

#### Cloud Offering
**Pattern:** Managed cloud service alongside self-hosted option.

```python
# Skyvern Cloud provides:
# - No infrastructure setup
# - Automatic scaling
# - Built-in anti-bot
# - Proxy network
```

**Lesson:** Cloud offering provides:
- Easy onboarding
- No infrastructure management
- Automatic scaling
- Better performance

**Applicability:** Consider cloud when:
- You don't want to manage infrastructure
- You need scaling
- You want better performance

---

## Cross-Repository Comparison

### Common Patterns

| Pattern | Browser-Use | Skyvern | Best Practice |
|---------|-------------|---------|---------------|
| **LLM Integration** | Unified interface | Unified interface | Always abstract LLM providers |
| **Browser Control** | CDP + Playwright | Playwright extension | Use Playwright for compatibility |
| **Event-Driven** | EventBus | Event system | Use events for loose coupling |
| **Error Handling** | Classified errors | Error handlers | Classify errors for specific handling |
| **Retry Logic** | Built-in retry | Built-in retry | Always retry with backoff |
| **Screenshot Capture** | Automatic | Automatic | Always capture screenshots for debugging |

### Different Approaches

| Aspect | Browser-Use | Skyvern | When to Use |
|--------|-------------|---------|-------------|
| **Task Definition** | Natural language | Workflow blocks | Simple vs complex tasks |
| **Element Selection** | DOM + XPath | Vision LLM | Stable vs changing UIs |
| **State Management** | In-memory | Database | Transient vs persistent |
| **Architecture** | Single service | Full-stack | Simple vs enterprise |
| **Deployment** | pip install | Docker Compose | Quick vs production |
| **Monitoring** | Basic logging | Dashboard | Simple vs comprehensive |

### Trade-offs

| Trade-off | Browser-Use Choice | Skyvern Choice | Lesson |
|----------|-------------------|----------------|--------|
| **Simplicity vs Power** | Simplicity | Power | Match tool to task complexity |
| **Speed vs Robustness** | Speed (DOM) | Robustness (Vision) | Vision is slower but more robust |
| **Setup vs Features** | Quick setup | More features | Consider onboarding time |
| **Cost vs Capability** | Lower cost | Higher cost | Budget vs requirements |
| **Flexibility vs Structure** | Flexible (natural language) | Structured (workflows) | Natural language vs visual builder |

---

## Applicable Patterns for Our Projects

### For Gmail Automation (Current Project)

**Recommended Approach:** Browser-Use

**Why:**
- Gmail login is a simple, repetitive task
- No need for workflow orchestration
- Natural language task definition fits perfectly
- Easy integration with existing Python code
- Lower cost and overhead

**Applicable Patterns:**
1. **Agent-first design** - Define task in natural language
2. **Profile management** - Use profiles for session persistence
3. **Error classification** - Classify login errors (wrong password, 2FA, CAPTCHA)
4. **Retry with backoff** - Retry failed logins
5. **Screenshot capture** - Capture screenshots for debugging
6. **Custom tools** - Add tools for Gmail-specific actions
7. **Token cost tracking** - Track LLM costs for 100 accounts

### For Complex Workflows (Future)

**Recommended Approach:** Skyvern

**When to use:**
- Multi-step workflows across multiple sites
- Need visual workflow builder for non-technical users
- Need database persistence
- Need comprehensive monitoring dashboard
- Need computer vision for complex UIs

**Applicable Patterns:**
1. **Workflow engine** - Declarative workflow definition
2. **Agent swarm** - Specialized agents for different tasks
3. **Vision-based detection** - Robust element selection
4. **Database persistence** - State recovery and history
5. **REST API** - External integration
6. **WebSocket** - Real-time updates
7. **Credential management** - Secure credential handling

---

## When to Use Each Approach

### Use Browser-Use When:

✅ **Task is simple and well-defined**
- Single action or short sequence
- Clear success criteria
- Minimal decision making

✅ **You want quick setup**
- pip install and go
- No infrastructure setup
- Minimal configuration

✅ **You prefer natural language**
- Define tasks in plain English
- No workflow building
- LLM handles decision logic

✅ **Budget is constrained**
- Lower infrastructure costs
- No database needed
- Simpler deployment

✅ **You're integrating into Python code**
- Native Python integration
- Easy to add to existing code
- Async/await patterns

### Use Skyvern When:

✅ **Task is complex and multi-step**
- Long workflows across multiple pages
- Conditional logic
- Loop structures

✅ **You need visual workflow builder**
- Non-technical users
- Drag-and-drop workflow creation
- Visual debugging

✅ **You need enterprise features**
- Database persistence
- User management
- Workflow versioning
- Comprehensive monitoring

✅ **You need computer vision**
- Complex UIs
- Changing layouts
- Image-based elements

✅ **You have a team**
- Backend developers (Python)
- Frontend developers (Node.js)
- Need collaboration features

---

## Key Takeaways

### 1. Match Tool to Task Complexity

**Lesson:** Don't over-engineer simple tasks.

- Simple tasks (Gmail login) → Browser-Use
- Complex workflows (multi-site scraping) → Skyvern

**Action:** Before choosing a tool, evaluate task complexity:
- Number of steps
- Decision complexity
- Need for persistence
- User technical level

### 2. Abstract LLM Providers

**Lesson:** Always abstract LLM providers for flexibility.

Both repositories do this well:
- Unified interface for multiple providers
- Easy provider switching
- Cost optimization through model routing

**Action:** In any LLM project:
- Create abstraction layer
- Support multiple providers
- Implement model routing
- Track costs

### 3. Event-Driven Architecture

**Lesson:** Event-driven architecture enables extensibility.

Browser-Use uses EventBus for:
- Browser events
- Agent events
- Custom events

**Action:** Use event-driven architecture when:
- Multiple independent components
- Need extensibility
- Want loose coupling
- Need observability

### 4. Error Classification

**Lesson:** Classify errors for specific handling.

Both repositories classify errors:
- Navigation errors
- Element not found
- Timeout errors
- LLM errors

**Action:** Always:
- Define error types
- Handle each type specifically
- Provide clear error messages
- Implement targeted retry logic

### 5. Screenshot Capture

**Lesson:** Screenshots are invaluable for debugging.

Both repositories automatically capture screenshots:
- On each step
- On errors
- On completion

**Action:** Always capture screenshots when:
- Debugging browser automation
- You need visual verification
- You want to show execution to users

### 6. Retry with Backoff

**Lesson:** Retry with exponential backoff is essential.

Both repositories implement retry:
- For network errors
- For LLM rate limits
- For temporary failures

**Action:** Always implement retry with backoff for:
- Network operations
- LLM API calls
- External service calls

### 7. Cost Tracking

**Lesson:** Track costs to optimize spending.

Browser-Use tracks:
- Input tokens
- Output tokens
- Cost per provider
- Total cost

**Action:** Always track costs when:
- Using LLM APIs
- You have budget constraints
- You want to optimize

### 8. Natural Language vs Workflows

**Lesson:** Natural language is faster, workflows are more structured.

Browser-Use: Natural language task definition
Skyvern: Visual workflow builder

**Action:** Choose based on:
- Task complexity
- User technical level
- Need for reusability
- Need for versioning

---

## Implementation Recommendations

### For Current Gmail Automation Project

**Phase 1: Setup (1-2 hours)**
```bash
# Install browser-use
pip install browser-use

# Get API key
# https://cloud.browser-use.com/new-api-key
```

**Phase 2: Simple Example (2-3 hours)**
```python
from browser_use import Agent, Browser, ChatBrowserUse
import asyncio

async def login_gmail(email, password, recovery_email):
    task = f"""
    Login to Gmail with:
    - Email: {email}
    - Password: {password}
    - Recovery email: {recovery_email}
    """
    
    browser = Browser()
    agent = Agent(
        task=task,
        llm=ChatBrowserUse(),
        browser=browser,
    )
    
    result = await agent.run()
    return result.is_successful()
```

**Phase 3: Integration (3-4 hours)**
- Integrate with existing monitor
- Add custom tools for Gmail-specific actions
- Implement error classification
- Add retry logic

**Phase 4: Production (2-3 hours)**
- Add rate limiting
- Implement batch processing
- Add comprehensive logging
- Test with 100 accounts

### For Future Complex Workflows

**When to consider Skyvern:**
- Multi-step workflows across multiple sites
- Need visual workflow builder
- Need database persistence
- Need comprehensive monitoring
- Need computer vision

**Implementation approach:**
1. Start with Skyvern Cloud (no infrastructure)
2. Evaluate if self-hosted is needed
3. Use workflow builder for complex tasks
4. Integrate with existing systems via REST API
5. Use WebSocket for real-time monitoring

---

## Conclusion

Both Browser-Use and Skyvern represent excellent approaches to AI-powered browser automation, but they serve different use cases:

**Browser-Use** is the right choice for:
- Simple, repetitive tasks
- Quick setup and deployment
- Natural language task definition
- Python-native integration
- Budget-constrained projects

**Skyvern** is the right choice for:
- Complex, multi-step workflows
- Enterprise-grade features
- Visual workflow building
- Computer vision needs
- Team collaboration

**Key Insight:** The most important lesson is to match the tool to the task. Don't over-engineer simple tasks with enterprise tools, and don't under-engineer complex tasks with simple tools.

For the current Gmail automation project, Browser-Use is the perfect fit: simple, repetitive login tasks that benefit from natural language task definition and Python-native integration.

---

**Next Steps:**
1. ✅ Research complete
2. ✅ Comparison documented
3. ✅ Lessons learned extracted
4. ⏭️ Integrate Browser-Use into Gmail automation
5. ⏭️ Test with single account
6. ⏭️ Scale to 100 accounts

---

*Document created: 2026-04-29*
*Source repositories analyzed: Browser-Use (91k stars), Skyvern (21.4k stars)*
*Total lessons extracted: 40+ patterns across architecture, browser control, error handling, performance, testing, integration, security, and deployment*
