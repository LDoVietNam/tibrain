# Skyvern Error Handling

**Repository**: Skyvern  
**Location**: `skyvern/exceptions.py` (1,137 lines)

---

## Overview

Skyvern's Error Handling system provides a comprehensive exception hierarchy with user-facing messages, browser connection error detection, and HTTP status code mapping. It ensures users receive actionable error messages while hiding internal implementation details.

---

## Exception Hierarchy

### Base Exceptions

```python
class SkyvernException(Exception):
    def __init__(self, message: str | None = None):
        self.message = message
        super().__init__(message)

class SkyvernClientException(SkyvernException):
    def __init__(self, message: str | None = None, status_code: int | None = None):
        self.status_code = status_code
        super().__init__(message)

class SkyvernHTTPException(SkyvernException):
    def __init__(self, message: str | None = None, status_code: int = status.HTTP_400_BAD_REQUEST):
        self.status_code = status_code
        super().__init__(message)
```

---

## Browser Connection Error Detection

**Location**: `exceptions.py` (lines 24-39)

### Error Patterns

```python
_BROWSER_CONNECTION_PATTERNS = (
    "connect_over_cdp",
    "WebSocket error",
    "WebSocket was closed",
    "ws connecting",
    "ws unexpected response",
    "ws error",
)

def _is_browser_connection_error(message: str) -> bool:
    return any(pattern in message for pattern in _BROWSER_CONNECTION_PATTERNS)
```

### User-Facing Message

```python
_BROWSER_CONNECTION_GUIDANCE = "Please try re-running. If this continues, contact support@skyvern.com."

def get_user_facing_exception_message(exception: Exception) -> str:
    if isinstance(exception, SkyvernException):
        return exception.message or str(exception)

    raw = str(exception)
    if _is_browser_connection_error(raw):
        return (
            f"Failed to connect to the browser session. "
            f"This is usually caused by high demand and is transient. {_BROWSER_CONNECTION_GUIDANCE}"
        )

    return f"Unexpected error: {exception}"
```

**Purpose**: Hide internal browser errors and provide actionable guidance

---

## HTTP Exceptions

### Rate Limit Exceeded

```python
class RateLimitExceeded(SkyvernHTTPException):
    def __init__(self, organization_id: str, max_requests: int, window_seconds: int):
        message = (
            f"Rate limit exceeded for organization {organization_id}. "
            f"Maximum {max_requests} requests per {window_seconds} seconds allowed."
        )
        super().__init__(message, status_code=status.HTTP_429_TOO_MANY_REQUESTS)
```

### Webhook Replay Error

```python
class WebhookReplayError(SkyvernHTTPException):
    def __init__(
        self,
        message: str | None = None,
        *,
        status_code: int = status.HTTP_400_BAD_REQUEST,
    ):
        super().__init__(message, status_code=status_code)
```

---

## Workflow Exceptions

### Disabled Block Execution

```python
class DisabledBlockExecutionError(SkyvernHTTPException):
    def __init__(self, message: str | None = None):
        super().__init__(message, status_code=status.HTTP_400_BAD_REQUEST)
```

### Invalid OpenAI Response

```python
class InvalidOpenAIResponseFormat(SkyvernException):
    def __init__(self, message: str | None = None):
        super().__init__(f"Invalid response format: {message}")
```

---

## Task Exceptions

### Task Not Found

```python
class TaskNotFound(SkyvernException):
    def __init__(self, task_id: str):
        super().__init__(f"Task {task_id} not found")
```

### Workflow Not Found

```python
class WorkflowNotFound(SkyvernException):
    def __init__(self, workflow_id: str):
        super().__init__(f"Workflow {workflow_id} not found")
```

### Workflow Run Not Found

```python
class WorkflowRunNotFound(SkyvernException):
    def __init__(self, workflow_run_id: str):
        super().__init__(f"Workflow run {workflow_run_id} not found")
```

---

## Browser Exceptions

### Browser Session Not Found

```python
class BrowserSessionNotFound(SkyvernException):
    def __init__(self, browser_session_id: str):
        super().__init__(f"Browser session {browser_session_id} not found")
```

### Browser Session Not Renewable

```python
class BrowserSessionNotRenewable(SkyvernException):
    def __init__(self, browser_session_id: str):
        super().__init__(f"Browser session {browser_session_id} is not renewable")
```

---

## Credential Exceptions

### Invalid Credential ID

```python
class InvalidCredentialId(SkyvernException):
    def __init__(self, credential_id: str):
        super().__init__(f"Invalid credential ID: {credential_id}")
```

### Failed To Fetch Secret

```python
class FailedToFetchSecret(SkyvernException):
    def __init__(self, secret_name: str):
        super().__init__(f"Failed to fetch secret: {secret_name}")
```

---

## Key Patterns

### 1. User-Facing Messages

**Pattern**: Provide actionable messages to users

**Benefits**:
- **Better UX**: Users know what to do
- **Support**: Clear guidance for common issues
- **Debugging**: Easier to diagnose issues

### 2. Browser Connection Error Detection

**Pattern**: Detect and hide browser connection errors

**Benefits**:
- **User experience**: Hide technical details
- **Actionable guidance**: Provide retry instructions
- **Security**: Don't expose internal URLs

### 3. HTTP Status Code Mapping

**Pattern**: Map exceptions to appropriate HTTP status codes

**Benefits**:
- **RESTful API**: Correct HTTP semantics
- **Client handling**: Clients can handle errors appropriately
- **Monitoring**: HTTP status codes for monitoring

### 4. Exception Hierarchy

**Pattern**: Organized exception hierarchy

**Benefits**:
- **Type safety**: Can catch specific exception types
- **Granular handling**: Different handling per exception type
- **Maintainability**: Easy to add new exceptions

---

## Testing Considerations

### Test Scenarios

1. **Browser connection errors** - Verify detection and messaging
2. **Rate limiting** - Verify rate limit exceptions
3. **User-facing messages** - Verify message quality
4. **Exception hierarchy** - Verify catch behavior
5. **HTTP status codes** - Verify correct mapping

---

## References

- **Exceptions**: `exceptions.py` (1,137 lines)
- **Errors Module**: `errors/errors.py`
