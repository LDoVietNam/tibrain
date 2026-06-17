---
tags: ["tibrain", "provider-gemini", "documentation", "router", "skill"]
scopes: ["providers", "tibrain"]
last_updated: 2026-05-22
---
# Gemini-API Pattern Analysis

## Repository: Gemini-API
**URL:** https://github.com/dsdanielpark/Gemini-API
**Location:** Z:\10_WORKPLACE\Ti\Ti-learning-lab\07_Repositories\router\Gemini-API

## Architecture Patterns

### 1. Automatic Cookie Extraction Pattern
**Pattern:** Browser-based cookie extraction with fallback
```python
def _set_cookies_automatically(self) -> None:
    if not self.auto_cookies and self.cookies is not None:
        return

    if self.auto_cookies:
        try:
            self._update_cookies_from_browser()
            if not self.cookies:
                raise ValueError("No cookies were loaded from the browser.")
        except Exception as e:
            raise Exception("Failed to extract cookies from browser.") from e
    else:
        # Fallback to auto_cookies
        self.auto_cookies = True
        self._update_cookies_from_browser()
```

**Key Learnings:**
- Automatic cookie extraction from browser
- Fallback mechanism for cookie loading
- browser_cookie3 package integration
- Multi-browser support

### 2. Multi-Browser Cookie Extraction Pattern
**Pattern:** Iterate through supported browsers
```python
def _update_cookies_from_browser(self) -> dict:
    for browser_fn in SUPPORTED_BROWSERS:
        try:
            cj = browser_fn(domain_name=".google.com")
            found_cookies = {cookie.name: cookie.value for cookie in cj}
            if len(found_cookies) >= 5:
                self.cookies = found_cookies
                break
        except Exception as e:
            print(e)
            continue
```

**Key Learnings:**
- Try multiple browsers sequentially
- Domain-specific cookie extraction
- Minimum cookie threshold (5 cookies)
- Graceful fallback on failure

### 3. Session ID and Nonce Extraction Pattern
**Pattern:** Parse session tokens from HTML
```python
def _set_sid_and_nonce(self):
    response = requests.get(f"{URLs.BASE_URL.value}/app", cookies=self.cookies)
    
    sid_match = re.search(r'"FdrFJe":"([\d-]+)"', response.text)
    nonce_match = re.search(r'"SNlM0e":"(.*?)"', response.text)
    
    if sid_match:
        self._sid = sid_match.group(1)
    if nonce_match:
        self._nonce = nonce_match.group(1)
    else:
        raise ValueError("Failed to parse SNlM0e nonce value")
```

**Key Learnings:**
- Extract session tokens from HTML response
- Regex pattern matching for tokens
- FdrFJe for session ID, SNlM0e for nonce
- Essential for request authentication

### 4. Request ID Generation Pattern
**Pattern:** Incremental request ID generation
```python
self._reqid = int("".join(random.choices(string.digits, k=7)))

def send_request(self, prompt, image=None):
    params = self._construct_params(self._sid)
    data = self._construct_payload(prompt, image, self._nonce)
    response = self.session.post(...)
    self._reqid += 100000  # Increment by 100000
```

**Key Learnings:**
- Random 7-digit request ID initialization
- Increment by 100000 for each request
- Used for request tracking
- Prevents request collision

### 5. Session Management Pattern
**Pattern:** Request session with headers and cookies
```python
def _initialize_session(self) -> requests.Session:
    session = requests.Session()
    session.headers.update(Headers.MAIN)
    if self.cookies:
        session.cookies.update(self.cookies)
    elif self.cookie_fp:
        session = self._set_cookies_from_file(session, self.cookie_fp)
    elif self.auto_cookies == True:
        self._set_cookies_automatically()
    
    self._set_sid_and_nonce()
    return session
```

**Key Learnings:**
- Session reuse for multiple requests
- Headers updated once at initialization
- Multiple cookie loading strategies
- Session ID extraction after cookie loading

### 6. URL Encoding Pattern
**Pattern:** URL-encoded parameters and payload
```python
def _construct_params(self, sid: str) -> str:
    return urllib.parse.urlencode({
        "bl": URLs.BOT_SERVER.value,
        "hl": os.environ.get("GEMINI_LANGUAGE", "en"),
        "_reqid": self._reqid,
        "rt": "c",
    })

def _construct_payload(self, prompt: str, image: Union[bytes, str], nonce: str) -> str:
    return urllib.parse.urlencode({
        "at": nonce,
        "f.req": json.dumps([None, json.dumps([...])]),
    })
```

**Key Learnings:**
- URL encoding for parameters
- Nested JSON encoding in payload
- Environment variable for language
- Nonce-based authentication

### 7. Response Parsing Pattern
**Pattern:** Multiple parser fallback
```python
def generate_custom_content(self, prompt: str, *custom_parsers) -> str:
    response_text, response_status_code = self.send_request(prompt)
    
    parser1 = ParseMethod1()
    parser2 = ParseMethod2()
    parsers = [parser1.parse, parser2.parse]
    
    for custom_parser in custom_parsers:
        if inspect.isclass(custom_parser):
            instance = custom_parser()
            parsers.append(instance.parse)
        elif callable(custom_parser):
            parsers.append(custom_parser)
    
    for parse in parsers:
        try:
            return parse(response_text)
        except Exception as e:
            continue
    
    return response_text
```

**Key Learnings:**
- Multiple parser strategies
- Graceful fallback on parse failure
- Support for custom parsers
- Return raw text on all failures

### 8. Candidate Collection Pattern
**Pattern:** Recursive stack-based data extraction
```python
@staticmethod
def collect_candidates(data: dict) -> list:
    collected = []
    stack = [data]
    
    while stack:
        current = stack.pop()
        
        if isinstance(current, dict):
            if "rcid" in current and "text" in current:
                collected.append(GeminiCandidate(**current))
            else:
                stack.extend(current.values())
        elif isinstance(current, list):
            stack.extend(current)
    
    return collected
```

**Key Learnings:**
- Stack-based traversal
- Handle nested dict/list structures
- Pattern matching for specific keys
- Collect all matching candidates

### 9. Cookie Filtering Pattern
**Pattern:** Target-specific cookie filtering
```python
if isinstance(self.target_cookies, list):
    filter_set = set(self.target_cookies)
elif self.target_cookies == "all":
    filter_set = WHOLE_COOKIES
else:
    filter_set = TARGET_COOKIES

self.cookies = {
    key: value for key, value in self.cookies.items() if key in filter_set
}
```

**Key Learnings:**
- Support for custom cookie lists
- "all" option for all cookies
- Default target cookies
- Dictionary comprehension for filtering

## Implementation Recommendations for Ti Router

### 1. Automatic Cookie Extraction
- Implement browser_cookie3 integration
- Support multiple browsers
- Add fallback mechanisms
- Set minimum cookie thresholds

### 2. Session Token Extraction
- Parse session tokens from HTML
- Use regex for pattern matching
- Handle token refresh
- Validate token validity

### 3. Request ID Management
- Implement incremental request IDs
- Random initialization
- Track request sequences
- Prevent collisions

### 4. Session Management
- Reuse sessions for multiple requests
- Update headers once at initialization
- Support multiple cookie loading strategies
- Extract session ID after cookie loading

### 5. Response Parsing
- Implement multiple parser strategies
- Add graceful fallback
- Support custom parsers
- Return raw text on failure

### 6. Cookie Filtering
- Support target-specific filtering
- Implement "all" option
- Use default target cookies
- Filter using set operations

## Priority
High - Apply these patterns for automatic cookie extraction and session management.
