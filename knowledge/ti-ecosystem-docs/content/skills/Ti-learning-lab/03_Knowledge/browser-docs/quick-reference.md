# Browser Automation Quick Reference

> **Source:** Browser-Use & Skyvern
> **Date:** 2026-04-29
> **Purpose:** Cheat sheet for common patterns and usage

---

## Browser-Use Quick Reference

### Gmail Integration

**Setup Gmail Service:**
```python
from browser_use.integrations.gmail import GmailService, register_gmail_actions
from browser_use import Agent, ChatBrowserUse, Tools

gmail_service = GmailService()
await gmail_service.authenticate()

tools = Tools()
register_gmail_actions(tools, gmail_service=gmail_service)
```

**Login with 2FA:**
```python
sensitive_data = {
    'gmail_email': email,
    'gmail_password': password,
    'recovery_email': recovery_email
}

agent = Agent(
    task='Login to Gmail. If 2FA appears, use get_recent_emails to find verification code.',
    llm=ChatBrowserUse(),
    tools=tools,
    sensitive_data=sensitive_data
)
```

### Custom Tools

**Create custom tool:**
```python
from browser_use import Tools

tools = Tools()

@tools.registry.action(description='Description')
async def custom_tool(param: str) -> ActionResult:
    return ActionResult(extracted_content='Result')
```

**Action filters:**
```python
@registry.action(description='Action', domains=['google.com'])
async def google_action(browser_session: BrowserSession):
    pass
```

### Form Filling

**Natural language form:**
```python
task = """
Go to https://example.com/form and fill:
- Name: John Doe
- Email: john@example.com
Then submit.
"""

agent = Agent(task=task, llm=llm)
```

### Sensitive Data

**Pass credentials securely:**
```python
sensitive_data = {'password': 'secret123'}

agent = Agent(
    task='Use password from sensitive_data',
    sensitive_data=sensitive_data
)
```

### Error Handling

**Retry with backoff:**
```python
agent = Agent(
    task=task,
    llm=llm,
    max_retries=3,
    retry_delay=1.0
)
```

---

## Skyvern Patterns Quick Reference

### Login Block → Browser-Use

**Skyvern:**
```json
{"block_type": "login", "url": "{{url}}", "parameter_keys": ["cred"]}
```

**Browser-Use:**
```python
task = "Login to {{url}} using credentials from sensitive_data"
```

### Navigation Block → Browser-Use

**Skyvern:**
```json
{"block_type": "navigation", "navigation_goal": "Fill form", "next_block_label": "next"}
```

**Browser-Use:**
```python
task = """
Step 1: Fill form
Step 2: Click Next
"""
```

### Extraction Block → Browser-Use

**Skyvern:**
```json
{"block_type": "extraction", "data_schema": {...}}
```

**Browser-Use:**
```python
task = "Extract account name, balance, and due date in structured format"
```

---

## Gmail Automation Quick Start

### 1. Install
```bash
pip install browser-use
export BROWSER_USE_API_KEY="your-key"
```

### 2. Setup
```python
from browser_use.integrations.gmail import GmailService, register_gmail_actions
from browser_use import Agent, ChatBrowserUse, Tools

gmail_service = GmailService()
await gmail_service.authenticate()

tools = Tools()
register_gmail_actions(tools, gmail_service=gmail_service)
```

### 3. Login
```python
sensitive_data = {
    'gmail_email': email,
    'gmail_password': password,
    'recovery_email': recovery_email
}

task = """
Login to Gmail at https://accounts.google.com using credentials from sensitive_data.
If 2FA appears, use get_recent_emails to find verification code.
"""

agent = Agent(task=task, llm=ChatBrowserUse(), tools=tools, sensitive_data=sensitive_data)
await agent.run()
```

---

## Common Tasks

### Read Recent Emails
```python
agent = Agent(
    task='Get recent emails with keyword "verification"',
    llm=llm,
    tools=tools
)
```

### Check 2FA Required
```python
@tools.registry.action(description='Check if 2FA is required')
async def check_2fa(browser_session: BrowserSession) -> ActionResult:
    # Check for 2FA prompt
    return ActionResult(extracted_content='2FA required: true/false')
```

### Extract Unread Count
```python
@tools.registry.action(description='Extract unread count')
async def extract_unread(browser_session: BrowserSession) -> ActionResult:
    # Extract unread count
    return ActionResult(extracted_content='Unread: 15')
```

---

## Error Classification

```python
class GmailLoginError:
    WRONG_PASSWORD = "wrong_password"
    ACCOUNT_DISABLED = "account_disabled"
    2FA_FAILED = "2fa_failed"
    CAPTCHA_REQUIRED = "captcha_required"
    NETWORK_ERROR = "network_error"

def classify_error(error: Exception) -> str:
    error_msg = str(error).lower()
    
    if 'wrong password' in error_msg:
        return GmailLoginError.WRONG_PASSWORD
    elif 'disabled' in error_msg:
        return GmailLoginError.ACCOUNT_DISABLED
    elif '2fa' in error_msg:
        return GmailLoginError.2FA_FAILED
    elif 'captcha' in error_msg:
        return GmailLoginError.CAPTCHA_REQUIRED
    elif 'network' in error_msg:
        return GmailLoginError.NETWORK_ERROR
    else:
        return "unknown"
```

---

## CLI Commands

### Browser-Use CLI
```bash
# Navigation
browser-use open <url>
browser-use state
browser-use click <index>
browser-use input <index> "text"

# Screenshot
browser-use screenshot page.png

# Session
browser-use close
```

### Setup
```bash
# Verify installation
browser-use doctor

# Setup credentials
browser-use cloud login <api-key>

# List profiles
browser-use profile list
```

---

## File Locations

### Browser-Use
- GmailService: `browser_use/integrations/gmail/service.py`
- Gmail Actions: `browser_use/integrations/gmail/actions.py`
- Custom Tools: `examples/custom-functions/`
- Form Filling: `examples/getting_started/02_form_filling.py`

### Skyvern
- Login Example: `skyvern/cli/skills/skyvern/examples/login-and-extract.json`
- Multi-Page Form: `skyvern/cli/skills/skyvern/examples/multi-page-form.json`

---

## Best Practices

### ✅ Do
- Use GmailService for Gmail automation
- Use sensitive_data for credentials
- Use action filters for security
- Implement error classification
- Test with single account first

### ❌ Don't
- Hardcode credentials in code
- Commit sensitive data to git
- Skip error handling
- Use complex workflows for simple tasks
- Skip testing before scaling

---

## Troubleshooting

### Gmail Authentication Failed
```python
# Use GmailGrantManager
from examples.integrations.gmail_2fa_integration import GmailGrantManager

grant_manager = GmailGrantManager()
await grant_manager.setup_oauth_credentials()
```

### 2FA Code Not Found
```python
# Test email reading
emails = await gmail_service.get_recent_emails(
    max_results=10,
    query='verification',
    time_filter='5m'
)
```

### Browser Timeout
```python
# Increase timeout
agent = Agent(task=task, llm=llm, max_steps=50)
```

---

## Documentation Links

- [Browser-Use Logic](./browser-use-logic.md) - Detailed Browser-Use components
- [Skyvern Patterns](./skyvern-patterns.md) - Skyvern workflow patterns
- [Gmail Implementation](./gmail-implementation-strategy.md) - Gmail automation guide
- [Skills List](./SKILLS_LIST.md) - Browser-Use built-in skills
- [Lessons Learned](./lesson.md) - Comprehensive lessons

---

*Quick reference created: 2026-04-29*
