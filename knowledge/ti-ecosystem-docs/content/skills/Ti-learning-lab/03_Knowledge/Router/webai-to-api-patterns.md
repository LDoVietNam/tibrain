# WebAI-to-API Pattern Analysis

## Repository: WebAI-to-API
**URL:** https://github.com/Amm1rr/WebAI-to-API
**Location:** Z:\10_WORKPLACE\Ti\Ti-learning-lab\07_Repositories\router\WebAI-to-API

## Architecture Patterns

### 1. Dual Server Architecture
**Pattern:** Hot-switching between WebAI and gpt4free servers
```python
# Main controller loop with hot-switching
while True:
    requested = shared_state["requested_mode"]
    if not current_process or (requested and requested != current_mode):
        # Gracefully stop current server
        if current_process and current_process.is_alive():
            stop_event.set()
            current_process.join(timeout=10)
        
        # Start new server in requested mode
        current_process = multiprocessing.Process(target=target_func, args=process_args)
        current_process.start()
```

**Key Learnings:**
- Multiprocessing for server isolation
- Graceful shutdown with timeout
- Shared state for mode switching
- Event-based coordination

### 2. Graceful Shutdown Pattern
**Pattern:** Signal handling and graceful shutdown
```python
def start_webai_server(host, port, reload, stop_event):
    signal.signal(signal.SIGINT, signal.SIG_IGN)
    if sys.platform == "win32":
        asyncio.set_event_loop_policy(asyncio.WindowsSelectorEventLoopPolicy())
    
    config = uvicorn.Config(webai_app, host=host, port=port, reload=reload, log_config=None)
    server = uvicorn.Server(config)
    
    def shutdown_monitor():
        stop_event.wait()
        server.should_exit = True
    
    monitor_thread = threading.Thread(target=shutdown_monitor, daemon=True)
    monitor_thread.start()
```

**Key Learnings:**
- Signal ignoring in child processes
- Event-based shutdown monitoring
- Thread-based monitoring (daemon)
- Windows event loop policy handling

### 3. Configuration Management
**Pattern:** TOML-based configuration with fallback
```python
try:
    import tomli
except ImportError:
    try:
        import tomllib as tomli
    except ImportError:
        tomli = None

def get_app_info():
    if not tomli:
        return "WebAI to API", "N/A (tomli not installed)"
    try:
        with open("pyproject.toml", "rb") as f:
            toml_data = tomli.load(f)
        poetry_data = toml_data.get("tool", {}).get("poetry", {})
        name = poetry_data.get("name", "WebAI-to-API").replace("-", " ").title()
        version = poetry_data.get("version", "N/A")
        return name, version
    except (FileNotFoundError, KeyError):
        return "WebAI-to-API", "N/A"
```

**Key Learnings:**
- Multiple import fallbacks
- Binary TOML reading
- Graceful degradation
- Poetry metadata extraction

### 4. Availability Checking
**Pattern:** Conditional mode availability
```python
webai_is_available = asyncio.run(init_gemini_client())
if webai_is_available:
    print(f"✅ WebAI-to-API mode is available")
else:
    print(f"⚠️ WebAI-to-API mode is not available")

if G4F_AVAILABLE:
    print(f"✅ gpt4free mode is available")
else:
    print(f"⚠️ gpt4free mode is not available")

initial_mode = "webai" if webai_is_available else "g4f" if G4F_AVAILABLE else None
```

**Key Learnings:**
- Async initialization checks
- Conditional availability
- Fallback mode selection
- User feedback on availability

### 5. Input Listener Pattern
**Pattern:** Non-blocking input handling
```python
def input_listener(shared_state):
    while True:
        try:
            choice = input()
            if choice == "1":
                shared_state["requested_mode"] = "webai"
            elif choice == "2":
                shared_state["requested_mode"] = "g4f"
        except (EOFError, KeyboardInterrupt):
            break

input_thread = threading.Thread(target=input_listener, args=(shared_state,), daemon=True)
input_thread.start()
```

**Key Learnings:**
- Thread-based input handling
- Shared state modification
- Daemon thread for cleanup
- Exception handling for graceful exit

## Implementation Recommendations for Ti Router

### 1. Server Hot-Switching
- Implement multiprocessing for provider switching
- Add graceful shutdown with timeout
- Use shared state for coordination

### 2. Graceful Shutdown
- Implement signal handling
- Add shutdown monitoring threads
- Handle Windows event loop policy

### 3. Configuration Management
- Use TOML for configuration
- Implement fallback imports
- Add graceful degradation

### 4. Availability Checking
- Add async initialization checks
- Implement conditional provider availability
- Provide user feedback

### 5. Input/Control Interface
- Implement non-blocking control interface
- Use shared state for coordination
- Add daemon threads for monitoring

## Priority
High - Apply these patterns for robust server management and graceful shutdown.
