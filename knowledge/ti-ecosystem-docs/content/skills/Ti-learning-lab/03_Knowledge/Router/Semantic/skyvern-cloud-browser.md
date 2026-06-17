# Skyvern Cloud Browser Integration

**Repository**: Skyvern  
**Location**: `skyvern/webeye/`  
**Key Files**:
- `browser_factory.py` - Browser session creation
- `browser_manager.py` - Browser lifecycle management
- `persistent_sessions_manager.py` - Persistent session management
- `default_persistent_sessions_manager.py` - Default session manager
- `cdp_download_interceptor.py` - CDP download interception
- `cdp_ports.py` - CDP port management
- `browser_state.py` - Browser state management
- `real_browser_state.py` - Real browser state

---

## Overview

Skyvern's Cloud Browser Integration manages browser sessions with stealth features, proxy rotation, CAPTCHA solving, and download interception. It supports both ephemeral and persistent sessions for different use cases.

---

## Browser Factory

**Location**: `webeye/browser_factory.py`

**Purpose**: Create browser sessions with configuration

**Features**:
- Playwright browser initialization
- Stealth configuration
- Proxy configuration
- Download directory setup
- CDP connection

---

## Browser Manager

**Location**: `webeye/browser_manager.py`

**Purpose**: Manage browser lifecycle

**Features**:
- Browser session creation
- Browser session cleanup
- Resource management
- Error handling

---

## Persistent Sessions

**Location**: `webeye/persistent_sessions_manager.py`

**Purpose**: Manage persistent browser sessions

**Features**:
- Session persistence across tasks
- Session renewal
- Session state tracking
- Session cleanup

### Default Persistent Sessions Manager

**Location**: `webeye/default_persistent_sessions_manager.py`

**Purpose**: Default implementation of persistent session manager

**Features**:
- In-memory session storage
- Basic session lifecycle
- Simple cleanup logic

---

## CDP Download Interceptor

**Location**: `webeye/cdp_download_interceptor.py`

**Purpose**: Intercept downloads via CDP

**Features**:
- Download interception
- File tracking
- Download completion detection
- Download timeout handling

---

## CDP Port Management

**Location**: `webeye/cdp_ports.py`

**Purpose**: Manage CDP ports for browser connections

**Features**:
- Port allocation
- Port cleanup
- Port conflict resolution

---

## Browser State

**Location**: `webeye/browser_state.py`

**Purpose**: Abstract browser state interface

**Features**:
- State abstraction
- Page access
- Element tree access
- Screenshot capture

### Real Browser State

**Location**: `webeye/real_browser_state.py`

**Purpose**: Real browser state implementation

**Features**:
- Playwright integration
- Real browser interaction
- State synchronization

---

## Stealth Features

**Purpose**: Evade bot detection

**Features**:
- User agent spoofing
- Webdriver property hiding
- Canvas fingerprint randomization
- WebGL fingerprint randomization
- Audio fingerprint randomization

---

## Proxy Rotation

**Purpose**: Rotate proxies for anonymity

**Features**:
- Proxy configuration
- Proxy rotation logic
- Proxy health checking
- Proxy error handling

---

## CAPTCHA Solving

**Purpose**: Solve CAPTCHAs automatically

**CAPTCHA Types**:
- Text CAPTCHA
- reCAPTCHA
- hCaptcha
- MTCaptcha
- FunCaptcha
- Cloudflare
- Other

**Integration**: Third-party CAPTCHA solving services

---

## Key Patterns

### 1. Session Persistence

**Pattern**: Reuse browser sessions across tasks

**Benefits**:
- Faster task execution
- Reduced overhead
- Maintains login state

### 2. Stealth Configuration

**Pattern**: Configure browser to evade detection

**Benefits**:
- Higher success rate
- Fewer blocks
- Better compatibility

### 3. Proxy Rotation

**Pattern**: Rotate proxies for anonymity

**Benefits**:
- Avoid IP blocks
- Geographic targeting
- Load distribution

### 4. Download Interception

**Pattern**: Intercept downloads via CDP

**Benefits**:
- Reliable download tracking
- Download completion detection
- File management

---

## Performance Optimizations

### 1. Session Reuse

**Impact**: Faster task execution

### 2. Resource Pooling

**Impact**: Reduced overhead

### 3. Lazy Initialization

**Impact**: Faster startup

---

## Testing Considerations

### Test Scenarios

1. **Session creation** - Verify browser initialization
2. **Session persistence** - Verify session reuse
3. **Stealth features** - Verify detection evasion
4. **Proxy rotation** - Verify proxy switching
5. **CAPTCHA solving** - Verify CAPTCHA handling
6. **Download interception** - Verify download tracking

### Test Commands

```bash
# Run browser tests
python -m pytest tests/unit/test_browser.py -v

# Run session tests
python -m pytest tests/unit/test_sessions.py -v
```

---

## References

- **Browser Factory**: `webeye/browser_factory.py`
- **Browser Manager**: `webeye/browser_manager.py`
- **Persistent Sessions**: `webeye/persistent_sessions_manager.py`
- **CDP Interceptor**: `webeye/cdp_download_interceptor.py`
- **Browser State**: `webeye/browser_state.py`
