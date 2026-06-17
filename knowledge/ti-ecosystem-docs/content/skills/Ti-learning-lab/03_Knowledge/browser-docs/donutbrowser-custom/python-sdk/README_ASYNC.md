# Donut Browser Python SDK - Async API

Async API for high-performance browser automation with AI agent integration.

## Installation

```bash
pip install aiohttp
```

## Quick Start

```python
import asyncio
from donut_browser_async import DonutBrowserAsync

async def main():
    async with DonutBrowserAsync(timeout=60, max_retries=3) as browser:
        # Create profile
        profile = await browser.create_profile(
            name="My Profile",
            browser="wayfern",
            tags=["automation"]
        )
        
        # Launch profile
        await browser.launch_profile(
            profile_id=profile["id"],
            url="https://example.com"
        )
        
        # Run AI task
        result = await browser.run_ai_task(
            profile_id=profile["id"],
            task_definition="Click the login button"
        )
        
        # Stop profile
        await browser.stop_profile(profile["id"])

asyncio.run(main())
```

## Key Features

### 1. Async Context Manager

Use the async context manager for automatic resource cleanup:

```python
async with DonutBrowserAsync() as browser:
    # Your code here
# Session automatically closed
```

### 2. Concurrent Batch Operations

Create or launch multiple profiles concurrently for 5-10x speedup:

```python
# Create multiple profiles concurrently
profiles = await browser.create_profiles_batch([
    {"name": "Profile 1", "tags": ["test"]},
    {"name": "Profile 2", "tags": ["test"]},
    {"name": "Profile 3", "tags": ["test"]},
])

# Launch multiple profiles concurrently
results = await browser.launch_profiles_batch(
    profile_ids=[p["id"] for p in profiles],
    headless=True
)
```

### 3. Connection Pooling

The async SDK uses connection pooling for efficient HTTP requests:
- 100 max concurrent connections
- 30 max connections per host
- Automatic connection cleanup

### 4. Retry Logic

Built-in retry logic for failed requests:

```python
browser = DonutBrowserAsync(
    timeout=30,
    max_retries=3  # Retry failed requests up to 3 times
)
```

## API Reference

### Profile Management

#### create_profile
```python
await browser.create_profile(
    name: str,
    browser: str = "wayfern",
    proxy_id: Optional[str] = None,
    tags: Optional[List[str]] = None
) -> Dict[str, Any]
```

#### list_profiles
```python
await browser.list_profiles() -> List[Dict[str, Any]]
```

#### get_profile
```python
await browser.get_profile(profile_id: str) -> Dict[str, Any]
```

#### delete_profile
```python
await browser.delete_profile(profile_id: str) -> Dict[str, Any]
```

### Profile Execution

#### launch_profile
```python
await browser.launch_profile(
    profile_id: str,
    url: Optional[str] = None,
    headless: bool = False
) -> Dict[str, Any]
```

#### stop_profile
```python
await browser.stop_profile(profile_id: str) -> Dict[str, Any]
```

### AI Task Management

#### run_ai_task
```python
await browser.run_ai_task(
    profile_id: str,
    task_definition: str,
    vision_enabled: bool = True,
    max_steps: int = 100
) -> Dict[str, Any]
```

#### get_agent_status
```python
await browser.get_agent_status(task_id: str) -> Dict[str, Any]
```

#### stop_agent
```python
await browser.stop_agent(task_id: str) -> Dict[str, Any]
```

### Computer Vision

#### detect_elements
```python
await browser.detect_elements(
    profile_id: str,
    filter_interactive: bool = True
) -> List[Dict[str, Any]]
```

### Batch Operations

#### create_profiles_batch
```python
await browser.create_profiles_batch(
    profiles: List[Dict[str, Any]]
) -> List[Dict[str, Any]]
```

#### launch_profiles_batch
```python
await browser.launch_profiles_batch(
    profile_ids: List[str],
    url: Optional[str] = None,
    headless: bool = False
) -> List[Dict[str, Any]]
```

## Performance Comparison

### Sync SDK
```
Create 10 profiles: ~10 seconds (sequential)
Launch 10 profiles: ~10 seconds (sequential)
```

### Async SDK
```
Create 10 profiles: ~1-2 seconds (concurrent)
Launch 10 profiles: ~1-2 seconds (concurrent)
```

**5-10x speedup for multi-account operations!**

## Error Handling

```python
from donut_browser_async import (
    DonutBrowserError,
    ProfileNotFoundError,
    TaskExecutionError
)

try:
    profile = await browser.create_profile(name="Test")
except ProfileNotFoundError as e:
    print(f"Profile not found: {e}")
except TaskExecutionError as e:
    print(f"Task execution failed: {e}")
except DonutBrowserError as e:
    print(f"General error: {e}")
```

## Advanced Usage

### Custom Session Configuration

```python
import aiohttp

# Create custom session
timeout = aiohttp.ClientTimeout(total=120)
connector = aiohttp.TCPConnector(limit=200, limit_per_host=50)

browser = DonutBrowserAsync(timeout=120, max_retries=5)
await browser._ensure_session()
# Override session if needed
```

### Parallel Task Execution

```python
async def process_profile(profile_id, task):
    async with DonutBrowserAsync() as browser:
        await browser.launch_profile(profile_id=profile_id)
        return await browser.run_ai_task(
            profile_id=profile_id,
            task_definition=task
        )

# Run tasks on multiple profiles concurrently
tasks = [
    process_profile(pid, "Task 1"),
    process_profile(pid2, "Task 2"),
]
results = await asyncio.gather(*tasks)
```

## Migration from Sync SDK

| Sync SDK | Async SDK |
|----------|-----------|
| `browser.create_profile()` | `await browser.create_profile()` |
| `browser.list_profiles()` | `await browser.list_profiles()` |
| `with DonutBrowser():` | `async with DonutBrowserAsync():` |
| Sequential batch | `create_profiles_batch()` (concurrent) |

## Best Practices

1. **Always use async context manager** for automatic resource cleanup
2. **Use batch operations** for multi-account workflows
3. **Set appropriate timeout** based on task complexity
4. **Handle exceptions** with specific error types
5. **Close sessions** when done (automatic with context manager)

## Requirements

- Python 3.7+
- aiohttp
- Donut Browser running locally

## License

Same as Donut Browser project.
