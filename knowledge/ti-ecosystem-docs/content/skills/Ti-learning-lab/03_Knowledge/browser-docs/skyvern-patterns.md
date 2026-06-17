---
tags: ["tibrain", "documentation", "websocket", "skill", "go"]
scopes: ["integration", "cli", "tibrain"]
last_updated: 2026-05-22
---
# Skyvern Workflow Patterns

> **Source:** Skyvern Repository (21.4k stars)
> **Location:** Z:\Ti\Ti-learning-lab\03_Knowledge\browser\skyvern
> **Date:** 2026-04-29
> **Focus:** 7 workflow patterns to adapt for Browser-Use

---

## Overview

Skyvern provides **7 workflow patterns** that can be adapted for Browser-Use. These patterns are more suitable for complex, multi-step workflows but need adaptation to work with Browser-Use's natural language approach.

---

## 1. Login Block Pattern

**Location:** `skyvern/cli/skills/skyvern/examples/login-and-extract.json`

**Purpose:** Declarative login definition with credential management

**Skyvern Pattern:**
```json
{
  "block_type": "login",
  "label": "login",
  "url": "{{portal_url}}",
  "title": "Login",
  "parameter_keys": ["login_credential"],
  "complete_criterion": "The account dashboard is visible and no login form is present."
}
```

**Features:**
- Declarative login definition
- Credential parameter passing
- Success criterion validation
- Automatic form filling

**Adaptation for Browser-Use:**
```python
task = """
Login to Gmail at https://accounts.google.com using credentials from sensitive_data.
Complete criterion: Gmail inbox is visible and logged in successfully.
"""
```

**Adaptation Steps:**
1. Convert JSON block to natural language task
2. Replace `parameter_keys` with `sensitive_data` dict
3. Convert `complete_criterion` to natural language success condition
4. Use Browser-Use's built-in form filling

---

## 2. Navigation Block Pattern

**Location:** `skyvern/cli/skills/skyvern/examples/multi-page-form.json`

**Purpose:** Natural language navigation with parameter interpolation

**Skyvern Pattern:**
```json
{
  "block_type": "navigation",
  "label": "personal_info",
  "url": "{{start_url}}",
  "title": "Personal Info",
  "navigation_goal": "Fill first name {{first_name}}, last name {{last_name}}, email {{email}}, then click Continue.",
  "next_block_label": "review_submit"
}
```

**Features:**
- Natural language navigation goals
- Parameter interpolation
- Sequential block execution
- Next block chaining

**Adaptation for Browser-Use:**
```python
task = """
Step 1: Go to {{start_url}}
Step 2: Fill first name {{first_name}}, last name {{last_name}}, email {{email}}
Step 3: Click Continue button
Step 4: Review entered data
Step 5: Submit the form
"""
```

**Adaptation Steps:**
1. Convert sequential blocks to numbered steps in task
2. Replace parameter interpolation with `sensitive_data` references
3. Use natural language for each step
4. Browser-Use will execute steps sequentially

---

## 3. Extraction Block Pattern

**Location:** `skyvern/cli/skills/skyvern/examples/login-and-extract.json`

**Purpose:** Structured data extraction with schema validation

**Skyvern Pattern:**
```json
{
  "block_type": "extraction",
  "label": "extract_summary",
  "title": "Extract Summary",
  "data_extraction_goal": "Extract account name, current balance, and next due date.",
  "data_schema": {
    "type": "object",
    "properties": {
      "account_name": {"type": "string"},
      "current_balance": {"type": "string"},
      "next_due_date": {"type": "string"}
    },
    "required": ["account_name", "current_balance"]
  }
}
```

**Features:**
- Structured data extraction
- JSON schema validation
- Required field enforcement
- Type checking

**Adaptation for Browser-Use:**
```python
task = """
Extract the following information from the page:
- Account name (required)
- Current balance (required)
- Next due date (optional)

Return the result in a structured format with these fields.
"""
```

**Adaptation Steps:**
1. Convert JSON schema to natural language description
2. Specify required fields in task
3. Browser-Use will return structured data
4. Add validation in post-processing if needed

**Applicable to Gmail:**
```python
task = """
After logging into Gmail, extract:
- Account email address (required)
- Number of unread emails (required)
- Subjects of 3 most recent emails (optional)

Return this information in a structured format.
"""
```

---

## 4. Credential Management

**Location:** Skyvern credential system

**Purpose:** Secure credential storage and retrieval

**Skyvern Pattern:**
```json
{
  "parameters": [
    {"parameter_type": "workflow", "key": "login_credential", "workflow_parameter_type": "credential_id"}
  ],
  "blocks": [
    {
      "block_type": "login",
      "parameter_keys": ["login_credential"]
    }
  ]
}
```

**Features:**
- Credential ID reference
- Automatic credential lookup
- Secure credential storage
- Multiple credential providers (Bitwarden, 1Password)

**Adaptation for Browser-Use:**
```python
# Use GmailGrantManager instead
from examples.integrations.gmail_2fa_integration import GmailGrantManager

grant_manager = GmailGrantManager()
await grant_manager.setup_oauth_credentials()

# Use sensitive_data for credentials
sensitive_data = {
    'gmail_email': email,
    'gmail_password': password
}
```

**Adaptation Steps:**
1. Use GmailGrantManager for OAuth credentials
2. Use `sensitive_data` dict for username/password
3. Browser-Use has better Gmail-specific credential management
4. No need for external credential providers for Gmail

---

## 5. Multi-Step Workflow

**Location:** Skyvern workflow engine

**Purpose:** Sequential block execution with parameter passing

**Skyvern Pattern:**
```json
{
  "blocks": [
    {
      "block_type": "navigation",
      "label": "personal_info",
      "next_block_label": "review_submit"
    },
    {
      "block_type": "navigation",
      "label": "review_submit",
      "next_block_label": "extract_confirmation"
    },
    {
      "block_type": "extraction",
      "label": "extract_confirmation",
      "next_block_label": null
    }
  ]
}
```

**Features:**
- Sequential execution
- Block chaining
- Parameter passing between blocks
- Workflow termination

**Adaptation for Browser-Use:**
```python
task = """
Step 1: Navigate to {{start_url}} and fill personal information
Step 2: Review the entered data
Step 3: Submit the form
Step 4: Extract confirmation number and status
"""
```

**Adaptation Steps:**
1. Convert sequential blocks to numbered steps
2. Use natural language for each step
3. Browser-Use will execute sequentially
4. Parameter passing via `sensitive_data`

**Applicable to Gmail:**
```python
task = """
Step 1: Navigate to https://accounts.google.com
Step 2: Fill login form with credentials from sensitive_data
Step 3: Click Next button
Step 4: If 2FA appears, use get_recent_emails to find verification code
Step 5: Enter the 2FA code
Step 6: Verify login success by checking for Gmail inbox
"""
```

---

## 6. Conditional Retry

**Location:** `skyvern/cli/skills/skyvern/examples/conditional-retry.json`

**Purpose:** Error type detection with custom retry strategies

**Skyvern Pattern:**
```json
{
  "blocks": [
    {
      "block_type": "navigation",
      "label": "attempt_action",
      "navigation_goal": "Perform the action",
      "error_handler": {
        "error_type": "ElementNotFound",
        "retry_strategy": "scroll_and_retry"
      }
    }
  ]
}
```

**Features:**
- Error type detection
- Custom retry strategies
- Scroll and retry
- Conditional execution

**Adaptation for Browser-Use:**
```python
# Browser-Use has built-in retry with exponential backoff
agent = Agent(
    task=task,
    llm=llm,
    max_retries=3,
    retry_delay=1.0  # Exponential backoff
)
```

**Adaptation Steps:**
1. Browser-Use retry is sufficient for most cases
2. Built-in exponential backoff
3. Automatic retry on network errors, LLM rate limits
4. No need for custom retry strategies in most cases

**When Custom Retry Needed:**
```python
# Implement custom retry in Python
async def custom_retry_login(max_attempts=3):
    for attempt in range(max_attempts):
        try:
            result = await login_gmail(...)
            return result
        except ElementNotFoundError:
            if attempt < max_attempts - 1:
                await asyncio.sleep(2 ** attempt)  # Exponential backoff
                continue
            raise
```

---

## 7. Parameter System

**Location:** Skyvern parameter system

**Purpose:** Parameter interpolation with type validation

**Skyvern Pattern:**
```json
{
  "parameters": [
    {"parameter_type": "workflow", "key": "start_url", "workflow_parameter_type": "string"},
    {"parameter_type": "workflow", "key": "first_name", "workflow_parameter_type": "string"},
    {"parameter_type": "workflow", "key": "last_name", "workflow_parameter_type": "string"}
  ],
  "blocks": [
    {
      "navigation_goal": "Fill first name {{first_name}}, last name {{last_name}}"
    }
  ]
}
```

**Features:**
- Parameter interpolation
- Type validation
- Required/optional parameters
- Default values

**Adaptation for Browser-Use:**
```python
# Use sensitive_data dict for parameter passing
sensitive_data = {
    'start_url': 'https://example.com',
    'first_name': 'John',
    'last_name': 'Doe'
}

task = """
Go to {{start_url}} and fill first name {{first_name}}, last name {{last_name}}
"""

agent = Agent(task=task, sensitive_data=sensitive_data)
```

**Adaptation Steps:**
1. Use `sensitive_data` dict for parameters
2. Use `{{parameter_name}}` syntax in task
3. Browser-Use will interpolate automatically
4. No type validation (rely on LLM)

**Applicable to Gmail:**
```python
sensitive_data = {
    'gmail_email': email,
    'gmail_password': password,
    'recovery_email': recovery_email
}

task = """
Login to Gmail using:
- Email: {{gmail_email}}
- Password: {{gmail_password}}
- Recovery email: {{recovery_email}}
"""
```

---

## Summary

| # | Pattern | Adaptation Needed | Complexity |
|---|---------|-------------------|------------|
| 1 | Login Block | Convert to natural language task | Low |
| 2 | Navigation Block | Convert to multi-step task | Low |
| 3 | Extraction Block | Convert to natural language extraction | Low |
| 4 | Credential Management | Use GmailGrantManager instead | Medium |
| 5 | Multi-Step Workflow | Implement as sequential task steps | Low |
| 6 | Conditional Retry | Browser-Use retry is sufficient | Low |
| 7 | Parameter System | Use sensitive_data dict | Low |

---

## When to Use Skyvern Patterns

**Use Skyvern patterns when:**
- You need visual workflow builder
- You have complex multi-site workflows
- You need database persistence
- You need enterprise-grade orchestration
- You have non-technical users

**Use Browser-Use adaptation when:**
- You prefer natural language tasks
- You want simpler setup
- You need Python-native integration
- You have technical users
- You want faster iteration

---

## Next Steps

- Read [browser-use-logic.md](./browser-use-logic.md) for Browser-Use components
- Read [gmail-implementation-strategy.md](./gmail-implementation-strategy.md) for Gmail implementation
- Read [quick-reference.md](./quick-reference.md) for cheat sheet

---

*Document created: 2026-04-29*
*Focus: Skyvern workflow patterns and Browser-Use adaptation*
