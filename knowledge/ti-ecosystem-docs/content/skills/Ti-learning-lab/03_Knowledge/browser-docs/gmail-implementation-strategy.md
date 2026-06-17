# Gmail Automation Implementation Strategy

> **Based on:** Browser-Use Gmail integration
> **Date:** 2026-04-29
> **Purpose:** Step-by-step implementation guide for Gmail automation

---

## Overview

This guide provides a **3-phase implementation strategy** for Gmail automation using Browser-Use's built-in Gmail integration.

---

## Phase 1: Use Browser-Use Gmail Integration (Immediate)

### Step 1: Setup Gmail Service

**Install browser-use:**
```bash
pip install browser-use
```

**Get API key:**
```bash
# Get from https://cloud.browser-use.com/new-api-key
export BROWSER_USE_API_KEY="your-key"
```

**Initialize Gmail service:**
```python
from browser_use.integrations.gmail import GmailService, register_gmail_actions
from browser_use import Agent, ChatBrowserUse, Tools
import asyncio

async def setup_gmail_service():
    # Initialize Gmail service
    gmail_service = GmailService()
    
    # Authenticate
    auth_success = await gmail_service.authenticate()
    
    if not auth_success:
        print("Gmail authentication failed")
        return None
    
    print("Gmail service ready")
    return gmail_service
```

**Register Gmail actions:**
```python
async def setup_tools(gmail_service):
    tools = Tools()
    register_gmail_actions(tools, gmail_service=gmail_service)
    return tools
```

---

### Step 2: Create Gmail Login Agent

**Basic login function:**
```python
async def login_gmail(email: str, password: str, recovery_email: str) -> bool:
    sensitive_data = {
        'gmail_email': email,
        'gmail_password': password,
        'recovery_email': recovery_email
    }
    
    task = """
    Login to Gmail at https://accounts.google.com using credentials from sensitive_data.
    
    Steps:
    1. Navigate to https://accounts.google.com
    2. Fill email field with gmail_email
    3. Fill password field with gmail_password
    4. Click Next button
    5. If 2FA appears:
       - Use get_recent_emails to find verification code from recovery_email
       - Extract the 2FA code from the email
       - Enter the 2FA code
    6. Verify login success by checking for Gmail inbox
    
    Return success if Gmail inbox is visible, otherwise return failure.
    """
    
    agent = Agent(
        task=task,
        llm=ChatBrowserUse(),
        tools=tools,
        sensitive_data=sensitive_data
    )
    
    history = await agent.run()
    return history.is_successful()
```

**Usage:**
```python
async def main():
    # Setup
    gmail_service = await setup_gmail_service()
    tools = await setup_tools(gmail_service)
    
    # Test with single account
    success = await login_gmail(
        email="test@example.com",
        password="password123",
        recovery_email="recovery@example.com"
    )
    
    print(f"Login success: {success}")
```

---

### Step 3: Integrate with Existing Monitor

**Connect to your monitor system:**
```python
from monitor import Monitor  # Your existing monitor

async def run_batch(accounts: list):
    monitor = Monitor()
    
    for account in accounts:
        monitor.start_account(account.email)
        
        try:
            success = await login_gmail(
                account.email,
                account.password,
                account.recovery_email
            )
            
            status = 'success' if success else 'failed'
            monitor.update_account_status(account.email, status)
            
        except Exception as e:
            monitor.update_account_status(account.email, f'error: {e}')
            print(f"Error for {account.email}: {e}")
    
    monitor.generate_report()
```

**Load accounts from Excel:**
```python
import pandas as pd

def load_accounts_from_excel(file_path: str) -> list:
    df = pd.read_excel(file_path)
    accounts = []
    
    for _, row in df.iterrows():
        accounts.append({
            'email': row['MAIL'],
            'password': row['PASS'],
            'recovery_email': row['RECOVER_MAIL']
        })
    
    return accounts

# Usage
accounts = load_accounts_from_excel('D:\\gmail\\gmail_accounts.xlsx')
await run_batch(accounts)
```

---

## Phase 2: Add Custom Tools (Enhancement)

### Custom Gmail Actions

**Add domain-specific actions:**
```python
@tools.registry.action(
    description='Check if Gmail 2FA is required on current page',
    domains=['accounts.google.com']
)
async def check_2fa_required(browser_session: BrowserSession) -> ActionResult:
    # Check for 2FA prompt
    # Return whether 2FA is required
    cdp_session = await browser_session.get_or_create_cdp_session()
    result = await cdp_session.cdp_client.send.Runtime.evaluate(
        params={'expression': 'document.querySelector("input[type=\'tel\']") !== null', 'returnByValue': True},
        session_id=cdp_session.session_id
    )
    
    is_required = result.get('result', {}).get('value', False)
    return ActionResult(
        extracted_content=f'2FA required: {is_required}',
        include_extracted_content_only_once=True
    )
```

**Extract Gmail unread count:**
```python
@tools.registry.action(
    description='Extract Gmail unread email count from inbox',
    domains=['mail.google.com']
)
async def extract_unread_count(browser_session: BrowserSession) -> ActionResult:
    cdp_session = await browser_session.get_or_create_cdp_session()
    result = await cdp_session.cdp_client.send.Runtime.evaluate(
        params={'expression': 'document.querySelector("[data-tooltip=\'Inbox\"]').getAttribute("aria-label")', 'returnByValue': True},
        session_id=cdp_session.session_id
    )
    
    aria_label = result.get('result', {}).get('value', '')
    # Extract number from aria-label like "Inbox (15)"
    import re
    match = re.search(r'\((\d+)\)', aria_label)
    unread_count = match.group(1) if match else "0"
    
    return ActionResult(
        extracted_content=f'Unread count: {unread_count}',
        include_extracted_content_only_once=True
    )
```

---

### Error Classification

**Define error types:**
```python
class GmailLoginError:
    WRONG_PASSWORD = "wrong_password"
    ACCOUNT_DISABLED = "account_disabled"
    2FA_FAILED = "2fa_failed"
    CAPTCHA_REQUIRED = "captcha_required"
    NETWORK_ERROR = "network_error"
    UNKNOWN_ERROR = "unknown_error"

def classify_error(error: Exception) -> str:
    error_msg = str(error).lower()
    
    if 'wrong password' in error_msg or 'incorrect password' in error_msg:
        return GmailLoginError.WRONG_PASSWORD
    elif 'disabled' in error_msg or 'suspended' in error_msg:
        return GmailLoginError.ACCOUNT_DISABLED
    elif '2fa' in error_msg or 'verification' in error_msg:
        return GmailLoginError.2FA_FAILED
    elif 'captcha' in error_msg or 'robot' in error_msg:
        return GmailLoginError.CAPTCHA_REQUIRED
    elif 'network' in error_msg or 'timeout' in error_msg:
        return GmailLoginError.NETWORK_ERROR
    else:
        return GmailLoginError.UNKNOWN_ERROR
```

**Error handling:**
```python
async def login_gmail_with_retry(email: str, password: str, recovery_email: str, max_retries=3) -> bool:
    for attempt in range(max_retries):
        try:
            success = await login_gmail(email, password, recovery_email)
            if success:
                return True
        except Exception as e:
            error_type = classify_error(e)
            print(f"Attempt {attempt + 1}: {error_type} - {e}")
            
            if error_type == GmailLoginError.WRONG_PASSWORD:
                # Don't retry wrong password
                return False
            elif error_type == GmailLoginError.ACCOUNT_DISABLED:
                # Don't retry disabled account
                return False
            elif attempt < max_retries - 1:
                # Retry with backoff
                await asyncio.sleep(2 ** attempt)
                continue
            else:
                return False
    
    return False
```

---

## Phase 3: Adapt Skyvern Patterns (Optimization)

### Multi-Step Task Structure

**Convert Skyvern workflow to Browser-Use task:**
```python
task = """
Step 1: Navigate to https://accounts.google.com
Step 2: Fill email field with {{gmail_email}}
Step 3: Fill password field with {{gmail_password}}
Step 4: Click Next button
Step 5: If 2FA prompt appears:
   - Use get_recent_emails to find verification code from {{recovery_email}}
   - Extract the 6-digit code from the email
   - Enter the 2FA code
   - Click Verify
Step 6: Verify login success by checking for Gmail inbox URL
Step 7: If successful, extract unread email count
Step 8: Return success status and unread count
"""
```

### Data Extraction Schema

**Add structured output:**
```python
from pydantic import BaseModel

class GmailLoginResult(BaseModel):
    success: bool
    email: str
    unread_count: int = 0
    error: str = ""

# In task, specify expected output
task += """
Return the result in this format:
{
  "success": true/false,
  "email": "logged-in-email",
  "unread_count": number,
  "error": "error-message-if-any"
}
"""
```

---

## Complete Implementation Example

```python
import asyncio
import pandas as pd
from browser_use.integrations.gmail import GmailService, register_gmail_actions
from browser_use import Agent, ChatBrowserUse, Tools
from monitor import Monitor

class GmailAutomator:
    def __init__(self):
        self.gmail_service = None
        self.tools = None
        self.monitor = Monitor()
    
    async def setup(self):
        """Initialize Gmail service and tools"""
        self.gmail_service = GmailService()
        auth_success = await self.gmail_service.authenticate()
        
        if not auth_success:
            raise Exception("Gmail authentication failed")
        
        self.tools = Tools()
        register_gmail_actions(self.tools, gmail_service=self.gmail_service)
        
        print("Gmail automator ready")
    
    async def login_account(self, email: str, password: str, recovery_email: str) -> dict:
        """Login to a single Gmail account"""
        self.monitor.start_account(email)
        
        sensitive_data = {
            'gmail_email': email,
            'gmail_password': password,
            'recovery_email': recovery_email
        }
        
        task = """
        Login to Gmail at https://accounts.google.com using credentials from sensitive_data.
        
        Steps:
        1. Navigate to https://accounts.google.com
        2. Fill email field with gmail_email
        3. Fill password field with gmail_password
        4. Click Next button
        5. If 2FA appears:
           - Use get_recent_emails to find verification code from recovery_email
           - Extract the 2FA code
           - Enter the 2FA code
        6. Verify login success by checking for Gmail inbox
        
        Return success status.
        """
        
        try:
            agent = Agent(
                task=task,
                llm=ChatBrowserUse(),
                tools=self.tools,
                sensitive_data=sensitive_data
            )
            
            history = await agent.run()
            success = history.is_successful()
            
            status = 'success' if success else 'failed'
            self.monitor.update_account_status(email, status)
            
            return {'email': email, 'success': success}
            
        except Exception as e:
            error_msg = str(e)
            self.monitor.update_account_status(email, f'error: {error_msg}')
            return {'email': email, 'success': False, 'error': error_msg}
    
    async def run_batch(self, excel_file: str):
        """Run batch automation for all accounts in Excel file"""
        accounts = load_accounts_from_excel(excel_file)
        
        results = []
        for account in accounts:
            result = await self.login_account(
                account['email'],
                account['password'],
                account['recovery_email']
            )
            results.append(result)
        
        self.monitor.generate_report()
        return results

# Usage
async def main():
    automator = GmailAutomator()
    await automator.setup()
    
    results = await automator.run_batch('D:\\gmail\\gmail_accounts.xlsx')
    
    # Print summary
    success_count = sum(1 for r in results if r['success'])
    print(f"Success: {success_count}/{len(results)}")

if __name__ == "__main__":
    asyncio.run(main())
```

---

## Testing Strategy

### Test 1: Single Account
```python
# Test with one account first
success = await login_gmail(
    email="test@example.com",
    password="test123",
    recovery_email="recovery@example.com"
)
print(f"Test result: {success}")
```

### Test 2: Small Batch (5 accounts)
```python
# Test with 5 accounts
test_accounts = accounts[:5]
for account in test_accounts:
    await login_account(account['email'], account['password'], account['recovery_email'])
```

### Test 3: Full Batch (100 accounts)
```python
# Run full batch
results = await automator.run_batch('D:\\gmail\\gmail_accounts.xlsx')
```

---

## Troubleshooting

### Gmail Authentication Failed
**Solution:** Use GmailGrantManager to setup OAuth credentials
```python
from examples.integrations.gmail_2fa_integration import GmailGrantManager

grant_manager = GmailGrantManager()
await grant_manager.setup_oauth_credentials()
```

### 2FA Code Not Found
**Solution:** Check recovery email is correct and emails are arriving
```python
# Test email reading
emails = await gmail_service.get_recent_emails(
    max_results=10,
    query='verification',
    time_filter='5m'
)
print(f"Found {len(emails)} verification emails")
```

### Browser Timeout
**Solution:** Increase timeout in Agent configuration
```python
agent = Agent(
    task=task,
    llm=llm,
    tools=tools,
    max_steps=50  # Increase max steps
)
```

---

## Next Steps

1. ✅ Install browser-use
2. ✅ Setup Gmail credentials using GmailGrantManager
3. ✅ Test with single account
4. ✅ Integrate with existing monitor
5. ✅ Add error classification
6. ✅ Test with small batch
7. ✅ Scale to full batch (100 accounts)

---

*Implementation guide created: 2026-04-29*
*Based on Browser-Use Gmail integration*
