# Browser-Use Available Logic

> **Source:** Browser-Use Repository (91k stars)
> **Location:** Z:\Ti\Ti-learning-lab\03_Knowledge\browser\browser-use
> **Date:** 2026-04-29
> **Focus:** 9 production-ready components for browser automation

---

## Overview

Browser-Use provides **9 production-ready components** that can be used immediately for browser automation, especially for Gmail automation.

---

## 1. GmailService

**Location:** `browser_use/integrations/gmail/service.py`

**Purpose:** Complete Gmail API integration with OAuth authentication

**Features:**
- OAuth2 authentication with Gmail API
- Token management and refresh
- Email reading with filtering
- 2FA code extraction
- Direct access token support

**Key Methods:**
```python
class GmailService:
    def __init__(self, credentials_file=None, token_file=None, access_token=None)
    async def authenticate() -> bool
    def is_authenticated() -> bool
    async def get_recent_emails(max_results=10, query='', time_filter='5m')
```

**Usage Example:**
```python
from browser_use.integrations.gmail import GmailService

# Initialize service
gmail_service = GmailService()
await gmail_service.authenticate()

# Get recent emails
emails = await gmail_service.get_recent_emails(
    max_results=5,
    query='verification',
    time_filter='5m'
)
```

**Status:** ✅ Production-ready, can be used immediately

---

## 2. GmailGrantManager

**Location:** `examples/integrations/gmail_2fa_integration.py`

**Purpose:** Credential management and OAuth grant flow

**Features:**
- Credential validation and setup
- Interactive OAuth grant flow
- Fallback re-authentication
- Error handling and recovery
- Token file management

**Key Methods:**
```python
class GmailGrantManager:
    def check_credentials_exist() -> bool
    def check_token_exists() -> bool
    def validate_credentials_format() -> tuple[bool, str]
    async def setup_oauth_credentials() -> bool
    async def test_authentication(gmail_service) -> tuple[bool, str]
    async def handle_authentication_failure(gmail_service, error_msg) -> bool
```

**Usage Example:**
```python
from examples.integrations.gmail_2fa_integration import GmailGrantManager

grant_manager = GmailGrantManager()

# Validate credentials
if not grant_manager.validate_credentials_format():
    await grant_manager.setup_oauth_credentials()

# Test authentication
gmail_service = GmailService()
auth_success, auth_message = await grant_manager.test_authentication(gmail_service)

# Handle failures
if not auth_success:
    await grant_manager.handle_authentication_failure(gmail_service, auth_message)
```

**Status:** ✅ Production-ready, can be used immediately

---

## 3. Gmail Actions

**Location:** `browser_use/integrations/gmail/actions.py`

**Purpose:** Pre-built tools for email reading and 2FA code extraction

**Features:**
- `get_recent_emails` action with keyword filtering
- Automatic authentication
- Time filtering (default: last 5 minutes)
- Full email content extraction
- Error handling and recovery

**Available Action:**
```python
@tools.registry.action(
    description='Get recent emails from the mailbox with a keyword to retrieve verification codes, OTP, 2FA tokens, magic links, or any recent email content.'
)
async def get_recent_emails(params: GetRecentEmailsParams) -> ActionResult
```

**Parameters:**
- `keyword` - Search keyword (e.g., "github", "verification")
- `max_results` - Number of emails (1-50, default: 3)

**Usage Example:**
```python
from browser_use.integrations.gmail import register_gmail_actions
from browser_use import Tools

tools = Tools()
register_gmail_actions(tools, gmail_service=gmail_service)

# Use in agent
agent = Agent(
    task='Search for 2FA verification codes in recent Gmail emails',
    llm=llm,
    tools=tools
)
```

**Status:** ✅ Production-ready, can be used immediately

---

## 4. Custom Tools Framework

**Location:** `examples/custom-functions/`

**Purpose:** Framework for adding domain-specific actions

**Features:**
- `@tools.registry.action` decorator
- ActionResult return type
- Parameter validation with Pydantic
- Easy integration with Agent

**Pattern:**
```python
from browser_use import Agent, Tools

tools = Tools()

@tools.registry.action(description='Description of what this tool does.')
async def custom_tool(param: str) -> ActionResult:
    # Your custom logic here
    return ActionResult(extracted_content='Result')

agent = Agent(task='Your task', llm=llm, tools=tools)
```

**Available Examples:**
- `2fa.py` - 2FA code handling with sensitive data
- `action_filters.py` - Domain-based action filtering
- `file_upload.py` - File upload handling
- `notification.py` - Notification systems
- `parallel_agents.py` - Parallel execution

**Status:** ✅ Production-ready, can be used immediately

---

## 5. Action Filters

**Location:** `examples/custom-functions/action_filters.py`

**Purpose:** Domain-based action scoping for security and performance

**Features:**
- Domain-based action filtering
- Conditional execution
- Security (passwords only on login pages)
- Reduces LLM decision fatigue

**Pattern:**
```python
@registry.action(description='Action description', domains=['google.com', '*.google.com'])
async def google_specific_action(browser_session: BrowserSession):
    # Only available on Google domains
    pass
```

**Benefits:**
- Prevents actions on wrong domains
- Security (sensitive actions only on specific pages)
- Reduces LLM decision fatigue
- Prevents stateful action mis-triggering

**Usage Example for Gmail:**
```python
@registry.action(description='Fill Gmail login form', domains=['accounts.google.com'])
async def fill_gmail_login(browser_session: BrowserSession, email: str, password: str):
    # Only available on Gmail login page
    # Fill login form
    pass
```

**Status:** ✅ Production-ready, can be used immediately

---

## 6. Form Filling Pattern

**Location:** `examples/getting_started/02_form_filling.py`

**Purpose:** Natural language form automation

**Features:**
- Natural language form description
- Automatic field detection
- Form submission
- Response verification

**Pattern:**
```python
task = """
Go to https://example.com/form and fill out the form with:
- Name: John Doe
- Email: john@example.com
- Phone: 555-1234
Then submit the form.
"""

agent = Agent(task=task, llm=llm)
await agent.run()
```

**Applicable to Gmail:**
```python
task = """
Go to https://accounts.google.com and fill out the login form with:
- Email: {email}
- Password: {password}
Handle any 2FA prompts if they appear.
"""
```

**Status:** ✅ Production-ready, can be used immediately

---

## 7. Authentication Pattern

**Location:** `browser_use/integrations/gmail/service.py`

**Purpose:** OAuth authentication flow and token management

**Features:**
- OAuth flow implementation
- Token refresh
- Direct access token support
- Credential validation

**OAuth Flow Pattern:**
```python
# File-based authentication
creds = Credentials.from_authorized_user_file(token_file, SCOPES)
if creds.expired and creds.refresh_token:
    creds.refresh(Request())
service = build('gmail', 'v1', credentials=creds)

# OAuth flow
flow = InstalledAppFlow.from_client_secrets_file(credentials_file, SCOPES)
creds = flow.run_local_server(port=8080)
```

**Direct Access Token Pattern:**
```python
gmail_service = GmailService(access_token="your-access-token")
# Skips file-based auth, uses token directly
```

**Benefits:**
- No file system access needed (with direct token)
- Faster authentication
- Suitable for cloud environments

**Status:** ✅ Production-ready, can be used immediately

---

## 8. Error Recovery Pattern

**Location:** `examples/integrations/gmail_2fa_integration.py`

**Purpose:** Robust error handling and recovery mechanisms

**Features:**
- Multiple fallback options
- Clear error messages
- Interactive recovery
- Token cleanup

**Pattern:**
```python
async def handle_authentication_failure(gmail_service, error_msg):
    # Option 1: Remove old token file
    if token_file.exists():
        token_file.unlink()
        # Retry authentication
    
    # Option 2: Re-setup credentials
    if not validate_credentials():
        await setup_oauth_credentials()
    
    # Option 3: Manual troubleshooting
    print('Manual troubleshooting steps...')
    # Retry with user confirmation
```

**Built-in Retry:**
```python
# Automatic retry on:
# - Network errors
# - LLM rate limits
# - Temporary failures

agent = Agent(
    task=task,
    llm=llm,
    max_retries=3,
    retry_delay=1.0  # Exponential backoff
)
```

**Status:** ✅ Production-ready, can be used immediately

---

## 9. Sensitive Data Handling

**Location:** `examples/custom-functions/2fa.py`

**Purpose:** Secure parameter passing for sensitive information

**Features:**
- Secure data passing to agent
- No logging of sensitive data
- Automatic TOTP generation
- Suitable for OTP/2FA

**Pattern:**
```python
sensitive_data = {'bu_2fa_code': secret_key}

agent = Agent(
    task='Use the 2FA code from bu_2fa_code',
    sensitive_data=sensitive_data
)
```

**Applicable to Gmail:**
```python
sensitive_data = {
    'gmail_email': email,
    'gmail_password': password,
    'recovery_email': recovery_email
}

agent = Agent(
    task='Login to Gmail using credentials from sensitive_data',
    sensitive_data=sensitive_data
)
```

**Benefits:**
- Secure credential passing
- No logging of sensitive data
- Automatic code generation (TOTP)

**Status:** ✅ Production-ready, can be used immediately

---

## Summary

| # | Component | Status | Key Benefit |
|---|-----------|--------|-------------|
| 1 | GmailService | ✅ Ready | Complete Gmail API integration |
| 2 | GmailGrantManager | ✅ Ready | Credential management and OAuth flow |
| 3 | Gmail Actions | ✅ Ready | 2FA code extraction |
| 4 | Custom Tools | ✅ Ready | Domain-specific actions |
| 5 | Action Filters | ✅ Ready | Domain-based security |
| 6 | Form Filling | ✅ Ready | Natural language form automation |
| 7 | Authentication | ✅ Ready | OAuth flow and token management |
| 8 | Error Recovery | ✅ Ready | Robust error handling |
| 9 | Sensitive Data | ✅ Ready | Secure credential passing |

---

## Next Steps

- Read [gmail-implementation-strategy.md](./gmail-implementation-strategy.md) for implementation guide
- Read [quick-reference.md](./quick-reference.md) for cheat sheet
- Read [skyvern-patterns.md](./skyvern-patterns.md) for Skyvern patterns

---

*Document created: 2026-04-29*
*Focus: Browser-Use production-ready components*
