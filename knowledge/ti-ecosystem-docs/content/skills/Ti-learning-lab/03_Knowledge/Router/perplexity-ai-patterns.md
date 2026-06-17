# Perplexity-AI Pattern Analysis

## Repository: perplexity-ai
**URL:** https://github.com/helallao/perplexity-ai
**Location:** Z:\10_WORKPLACE\Ti\Ti-learning-lab\07_Repositories\router\perplexity-ai

## Architecture Patterns

### 1. Session Initialization Pattern
**Pattern:** Session initialization with browser impersonation
```python
def __init__(self, cookies={}):
    self.session = requests.Session(
        headers=DEFAULT_HEADERS.copy(),
        cookies=cookies,
        impersonate="chrome",
    )
    
    # Flags for account and query management
    self.own = bool(cookies)
    self.copilot = 0 if not cookies else float("inf")
    self.file_upload = 0 if not cookies else float("inf")
    
    # Initialize session
    self.session.get(ENDPOINT_AUTH_SESSION)
```

**Key Learnings:**
- curl_cffi for browser impersonation
- Session initialization with auth endpoint
- Query limit tracking
- File upload limit tracking

### 2. Account Creation Pattern
**Pattern:** Email-based account creation with email verification
```python
def create_account(self, cookies):
    while True:
        try:
            emailnator_cli = Emailnator(cookies)
            
            # Initiate account creation
            resp = self.session.post(
                ENDPOINT_AUTH_SIGNIN,
                data={
                    "email": emailnator_cli.email,
                    "csrfToken": self.session.cookies.get_dict()["next-auth.csrf-token"].split("%")[0],
                    "callbackUrl": "https://www.perplexity.ai/",
                    "json": "true",
                },
            )
            
            # Wait for sign-in email
            new_msgs = emailnator_cli.reload(
                wait_for=lambda x: x["subject"] == "Sign in to Perplexity",
                timeout=20,
            )
            
            if new_msgs:
                break
        except Exception:
            pass
    
    # Extract sign-in link and complete
    msg = emailnator_cli.get(func=lambda x: x["subject"] == "Sign in to Perplexity")
    new_account_link = self.signin_regex.search(emailnator_cli.open(msg["messageID"])).group(1)
    self.session.get(new_account_link)
    
    # Update limits
    self.copilot = 5
    self.file_upload = 10
```

**Key Learnings:**
- Email-based account creation
- CSRF token extraction
- Email waiting with timeout
- Sign-in link extraction via regex
- Query limit initialization

### 3. File Upload Pattern
**Pattern:** Two-stage file upload (metadata + S3 upload)
```python
# Stage 1: Get upload metadata
file_upload_info = (
    self.session.post(
        ENDPOINT_UPLOAD_URL,
        params={"version": "2.18", "source": "default"},
        json={
            "content_type": file_type,
            "file_size": sys.getsizeof(file),
            "filename": filename,
            "force_image": False,
            "source": "default",
        },
    )
).json()

# Stage 2: Upload to S3
mp = CurlMime()
for key, value in file_upload_info["fields"].items():
    mp.addpart(name=key, data=value)
mp.addpart(
    name="file",
    content_type=file_type,
    filename=filename,
    data=file,
)

upload_resp = self.session.post(file_upload_info["s3_bucket_url"], multipart=mp)

# Extract uploaded URL
if "image/upload" in file_upload_info["s3_object_url"]:
    uploaded_url = re.sub(
        r"/private/s--.*?--/v\d+/user_uploads/",
        "/private/user_uploads/",
        upload_resp.json()["secure_url"],
    )
else:
    uploaded_url = file_upload_info["s3_object_url"]
```

**Key Learnings:**
- Two-stage upload process
- Metadata request before upload
- S3 direct upload with signed fields
- URL transformation for images
- curl_cffi multipart support

### 4. Query Validation Pattern
**Pattern:** Comprehensive parameter validation
```python
def search(self, query, mode="auto", model=None, sources=["web"], files={}, stream=False, language="en-US", follow_up=None, incognito=False):
    # Validate mode
    assert mode in ["auto", "pro", "reasoning", "deep research"], "Invalid search mode."
    
    # Validate model for mode
    assert model in {
        "auto": [None],
        "pro": [None, "sonar", "gpt-5.2", "claude-4.5-sonnet", "grok-4.1"],
        "reasoning": [None, "gpt-5.2-thinking", "claude-4.5-sonnet-thinking", "gemini-3.0-pro", "kimi-k2-thinking", "grok-4.1-reasoning"],
        "deep research": [None],
    }[mode] if self.own else True, "Invalid model for the selected mode."
    
    # Validate sources
    assert all([source in ("web", "scholar", "social") for source in sources]), "Invalid sources."
    
    # Validate query limits
    assert self.copilot > 0 if mode in ["pro", "reasoning", "deep research"] else True, "No remaining pro queries."
    
    # Validate file upload limits
    assert self.file_upload - len(files) >= 0 if files else True, "File upload limit exceeded."
    
    # Update counters
    self.copilot = self.copilot - 1 if mode in ["pro", "reasoning", "deep research"] else self.copilot
    self.file_upload = self.file_upload - len(files) if files else self.file_upload
```

**Key Learnings:**
- Mode-based validation
- Model-per-mode validation
- Source whitelist validation
- Query limit checking
- File upload limit checking
- Counter decrement

### 5. Model Preference Mapping Pattern
**Pattern:** Nested dictionary for model mapping
```python
json_data = {
    "query_str": query,
    "params": {
        "attachments": uploaded_files + follow_up["attachments"] if follow_up else uploaded_files,
        "frontend_context_uuid": str(uuid4()),
        "frontend_uuid": str(uuid4()),
        "is_incognito": incognito,
        "language": language,
        "last_backend_uuid": follow_up["backend_uuid"] if follow_up else None,
        "mode": "concise" if mode == "auto" else "copilot",
        "model_preference": {
            "auto": {None: "turbo"},
            "pro": {
                None: "pplx_pro",
                "sonar": "experimental",
                "gpt-5.2": "gpt52",
                "claude-4.5-sonnet": "claude45sonnet",
                "grok-4.1": "grok41nonreasoning",
            },
            "reasoning": {
                None: "pplx_reasoning",
                "gpt-5.2-thinking": "gpt52_thinking",
                "claude-4.5-sonnet-thinking": "claude45sonnetthinking",
                "gemini-3.0-pro": "gemini30pro",
                "kimi-k2-thinking": "kimik2thinking",
                "grok-4.1-reasoning": "grok41reasoning",
            },
            "deep research": {None: "pplx_alpha"},
        }[mode][model],
        "source": "default",
        "sources": sources,
        "version": "2.18",
    },
}
```

**Key Learnings:**
- Nested dictionary mapping
- UUID generation for context
- Mode-based model selection
- Version parameter
- Follow-up context support

### 6. SSE Streaming Pattern
**Pattern:** Server-Sent Events with nested JSON parsing
```python
def stream_response(resp):
    for chunk in resp.iter_lines(delimiter=b"\r\n\r\n"):
        content = chunk.decode("utf-8")
        
        if content.startswith("event: message\r\n"):
            try:
                content_json = json.loads(content[len("event: message\r\ndata: ") :])
                
                # Parse nested 'text' field
                if "text" in content_json and content_json["text"]:
                    try:
                        text_parsed = json.loads(content_json["text"])
                        # Extract answer from FINAL step
                        if isinstance(text_parsed, list):
                            for step in text_parsed:
                                if step.get("step_type") == "FINAL":
                                    final_content = step.get("content", {})
                                    if "answer" in final_content:
                                        answer_data = json.loads(final_content["answer"])
                                        content_json["answer"] = answer_data.get("answer", "")
                                        content_json["chunks"] = answer_data.get("chunks", [])
                                        break
                        content_json["text"] = text_parsed
                    except (json.JSONDecodeError, TypeError, KeyError):
                        pass
                
                chunks.append(content_json)
                yield chunks[-1]
            except (json.JSONDecodeError, KeyError):
                continue
        
        elif content.startswith("event: end_of_stream\r\n"):
            return
```

**Key Learnings:**
- SSE event parsing
- Nested JSON extraction
- Step-based answer extraction
- FINAL step detection
- Graceful error handling

### 7. Query Limit Management Pattern
**Pattern:** Per-mode query limits
```python
# Initialize limits
self.copilot = 0 if not cookies else float("inf")  # Own account = unlimited
self.file_upload = 0 if not cookies else float("inf")

# Account creation sets limits
self.copilot = 5
self.file_upload = 10

# Decrement on usage
self.copilot = self.copilot - 1 if mode in ["pro", "reasoning", "deep research"] else self.copilot
self.file_upload = self.file_upload - len(files) if files else self.file_upload

# Validate before use
assert self.copilot > 0 if mode in ["pro", "reasoning", "deep research"] else True, "No remaining pro queries."
```

**Key Learnings:**
- Unlimited for own accounts
- Fixed limits for generated accounts
- Per-mode decrement
- Pre-use validation

## Implementation Recommendations for Ti Router

### 1. Session Management
- Use curl_cffi for browser impersonation
- Initialize session with auth endpoint
- Track query and upload limits
- Support both own and generated accounts

### 2. Account Creation
- Implement email-based account creation
- Add CSRF token extraction
- Implement email waiting with timeout
- Extract sign-in links via regex
- Initialize limits for new accounts

### 3. File Upload
- Implement two-stage upload process
- Request metadata before upload
- Support S3 direct upload
- Transform URLs for images
- Use curl_cffi for multipart

### 4. Query Validation
- Implement mode-based validation
- Add model-per-mode validation
- Validate source whitelist
- Check query limits
- Check file upload limits
- Decrement counters

### 5. Model Mapping
- Implement nested dictionary mapping
- Generate UUIDs for context
- Support mode-based model selection
- Add version parameter
- Support follow-up context

### 6. SSE Streaming
- Implement SSE event parsing
- Parse nested JSON
- Extract answers from steps
- Detect FINAL step
- Add graceful error handling

### 7. Limit Management
- Track per-mode query limits
- Support unlimited for own accounts
- Set fixed limits for generated accounts
- Decrement on usage
- Validate before use

## Priority
High - Apply these patterns for file upload, query validation, and limit management.
