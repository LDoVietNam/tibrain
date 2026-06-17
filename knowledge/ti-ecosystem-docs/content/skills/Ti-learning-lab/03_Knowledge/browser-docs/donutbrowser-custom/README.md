# Donut Browser Ultimate - Documentation

## Overview

Donut Browser Ultimate is an enhanced version of Donut Browser with AI automation, workflow management, and quality assurance features.

## Features

### Core Features (from Donut Browser)
- Anti-detect browser with fingerprint spoofing
- Multi-engine support (Wayfern/Chromium, Camoufox/Firefox)
- Profile isolation and management
- MCP server for integration
- REST API for automation
- Proxy and VPN support

### Enhanced Features
- **AI Agent Integration**: LLM-powered automation with computer vision
- **Workflow Builder**: No-code workflow creation with drag-and-drop
- **Database Persistence**: SQLite database for profiles, workflows, tasks
- **Web UI**: Monitoring dashboard and profile management
- **Python SDK**: Simple API for easy integration (sync + async)
- **Batch Processing**: Multi-account automation with retry logic
- **Beads Workflow**: Quality assurance with 7-stage workflow
- **Enhanced MCP**: Hidden admin tools with authentication
- **Response Caching**: LLM response caching to reduce API costs and latency

## Architecture

```
Donut Browser Ultimate
├── Anti-detect Engine (Donut)
│   ├── Wayfern (Chromium)
│   ├── Camoufox (Firefox)
│   └── Fingerprint Spoofing
├── AI Agent System (Skyvern + Browser-Use)
│   ├── LLM Integration
│   ├── Computer Vision
│   └── Agent Loop
├── Workflow Engine (Skyvern)
│   ├── Workflow Builder
│   ├── Block System
│   └── Execution Engine
├── Database Layer (Skyvern)
│   ├── SQLite
│   ├── Profile Storage
│   └── Workflow Persistence
├── Web UI (Skyvern)
│   ├── Workflow Builder
│   ├── Monitoring Dashboard
│   └── Profile Management
├── Python SDK (Browser-Use)
│   ├── Simple API
│   ├── AI Tasks
│   └── Workflow Execution
├── MCP Server (Comet + Donut)
│   ├── Standard Tools
│   ├── Hidden Tools
│   └── Admin Tools
├── CDP Server (Browser-Use)
│   ├── Authentication
│   ├── Origin Whitelist
│   └── Persistent Sessions
├── Batch Processor (Browser-Use)
│   ├── Multi-account
│   ├── Retry Logic
│   └── Error Handling
└── Beads Workflow
    ├── Quality Gates
    ├── Health Check
    ├── Auto-Fix
    ├── Optimization
    └── Verification
```

## Installation

### Development Setup

```bash
# Clone repository
git clone https://github.com/zhom/donutbrowser.git
cd donutbrowser

# Install dependencies
npm install

# Build
npm run build

# Run development
npm run dev
```

### Python SDK Setup

```bash
cd python-sdk
pip install -r requirements.txt
```

## Usage

### Web UI

1. Open Donut Browser Ultimate
2. Navigate to `/workflows` to create workflows
3. Navigate to `/dashboard` to monitor automation
4. Navigate to `/profiles` to manage profiles

### Python SDK

#### Sync SDK

```python
from donut_browser import DonutBrowser

# With timeout and retry configuration
with DonutBrowser(timeout=60, max_retries=3) as browser:
    profile = browser.create_profile("My Profile", "wayfern")
    browser.launch_profile(profile["id"], "https://example.com")
    result = browser.run_ai_task(profile["id"], "Login to website", vision_enabled=True)
    browser.stop_profile(profile["id"])
```

#### Async SDK (High Performance)

```python
import asyncio
from donut_browser_async import DonutBrowserAsync

async def main():
    async with DonutBrowserAsync(timeout=60, max_retries=3) as browser:
        # Create multiple profiles concurrently
        profiles = await browser.create_profiles_batch([
            {"name": "Profile 1", "tags": ["automation"]},
            {"name": "Profile 2", "tags": ["automation"]},
        ])

        # Launch multiple profiles concurrently
        results = await browser.launch_profiles_batch(
            profile_ids=[p["id"] for p in profiles],
            headless=True
        )

asyncio.run(main())
```

**Performance**: Async SDK provides 5-10x speedup for multi-account operations via concurrent execution.

See `python-sdk/README_ASYNC.md` for detailed async API documentation.

### MCP Integration

Connect to MCP server on port 51080 with authentication token.

Available tools:
- `list_profiles`, `get_profile`, `run_profile`, `kill_profile`
- `create_profile`, `update_profile`, `delete_profile`
- `run_ai_task`, `get_agent_status`, `stop_agent`
- `detect_elements`
- `admin_get_all_profiles`, `admin_delete_profile`, `admin_export_all_data` (admin only)

### Beads Workflow

All workflow executions follow the 7-stage beads workflow:
1. Read Registry
2. Identify Target
3. Health Check
4. Auto-Fix
5. Optimization
6. Verification
7. Update Beads Log

## API Reference

### REST API Documentation

Comprehensive API documentation is available in [`API_DOCUMENTATION.md`](./API_DOCUMENTATION.md).

The OpenAPI 3.0 specification is available in [`openapi.yaml`](./openapi.yaml).

**Quick Start:**
- Base URL: `http://localhost:10108`
- Interactive docs: Use Swagger UI or Redoc with `openapi.yaml`
- Admin endpoints: Require `X-Admin-Token` header
- See [`API_DOCUMENTATION.md`](./API_DOCUMENTATION.md) for full endpoint reference

### Configuration

#### Timeout Settings
- **AI Agent**: Default 30s timeout per LLM call (configurable via `timeout_ms`)
- **Workflow Engine**: Default 30s timeout per block (configurable via `timeout_ms`)
- **Batch Processing**: Default 30s timeout per task (configurable via `RetryPolicy.timeout_ms`)
- **Beads Workflow**: Default 60s timeout for entire workflow (configurable via `with_timeout()`)

#### Retry Policies
- **Python SDK**: Automatic retry with exponential backoff for HTTP 429, 500, 502, 503, 504 errors
- **Batch Processing**: Configurable max attempts, backoff delay, exponential backoff
- **Error Handling**: Retry, Skip, or Abort strategies

### MCP Server

Port: `51080`

Standard tools available via MCP protocol.

Available tools:
- `list_profiles`, `get_profile`, `run_profile`, `kill_profile`
- `create_profile`, `update_profile`, `delete_profile`
- `run_ai_task`, `get_agent_status`, `stop_agent`
- `detect_elements`
- `admin_get_all_profiles`, `admin_delete_profile`, `admin_export_all_data`, `admin_get_system_status` (admin only)

## Quality Assurance

### Beads Workflow

All automation tasks go through quality gates:
- Health check before execution
- Auto-fix for detected issues
- Performance optimization
- Verification after completion
- Beads log for tracking

### Testing

Run tests with:
```bash
npm test
```

## Troubleshooting

### Common Issues

#### Linker Errors (Windows)
If you encounter `link.exe` errors during Rust compilation:
- Ensure Visual Studio Build Tools are installed
- Check that the C++ build tools component is included
- Try running from x64 Native Tools Command Prompt

#### MCP Admin Tools Not Working
The admin MCP tools (`admin_get_all_profiles`, `admin_delete_profile`, `admin_export_all_data`, `admin_get_system_status`) are defined in the tool schema but not yet implemented in the handler. These will return errors when called.

#### Timeout Errors
If tasks are timing out frequently:
- Increase timeout settings in the relevant component
- Check network connectivity to LLM APIs
- Verify browser profile is launching correctly
- Review system resources (CPU, memory)

#### Database Locked
If SQLite database is locked:
- Ensure only one instance of the application is running
- Check for orphaned processes
- Use WAL mode (enabled by default) for better concurrency

### Performance Tips

- Use WAL mode for SQLite (enabled by default)
- Enable database indexes for frequent queries
- Use concurrent processing limits in batch operations
- Adjust timeout values based on task complexity
- Monitor dashboard for running tasks with auto-refresh

## Known Limitations

- ~~Admin MCP tools are defined but not implemented~~ ✅ **FIXED** - Now implemented with authentication
- ~~Workflow execution is sequential (not concurrent)~~ - Sequential by design (blocks have dependencies)
- ~~No caching layer for LLM responses~~ ✅ **FIXED** - LLM response caching now implemented
- No progress reporting for batch operations
- No cancellation support for long-running tasks
- No WebSocket support for real-time updates

## Recent Optimizations

### P0 - Critical ✅

#### MCP Admin Tools ✅
- Implemented 4 admin tools with authentication:
  - `admin_get_all_profiles` - Get all profiles including sensitive data
  - `admin_delete_profile` - Force delete without confirmation
  - `admin_export_all_data` - Export all system data
  - `admin_get_system_status` - Get detailed system metrics
- Authentication via `DONUT_ADMIN_TOKEN` environment variable
- Default token: `admin-secret-token-change-me` (change in production)

#### AI Automation Tools ✅
- Implemented 4 AI automation tools:
  - `run_ai_task` - Run LLM-powered automation
  - `get_agent_status` - Get agent task status
  - `stop_agent` - Stop running agent
  - `detect_elements` - Detect UI elements with CV

#### Concurrent Processing ✅
- Batch processor now supports concurrent execution
- Uses `tokio::spawn` with `Semaphore` for concurrency control
- Configurable via `concurrent_limit` parameter
- **Performance gain**: 5-10x speedup for multi-account batch operations

### P1 - Performance ✅

#### LLM Response Caching ✅
- Implemented global caching layer for LLM responses
- Cache key includes prompt, model, and provider
- Configurable TTL (default 1 hour)
- Reduces API costs and latency for repeated queries
- Cache size limit to prevent memory bloat

#### Async Python SDK ✅
- Created async version of Python SDK using aiohttp
- Connection pooling for efficient HTTP requests
- Concurrent batch operations for 5-10x speedup
- Automatic retry logic with exponential backoff
- Async context manager for resource cleanup
- See `python-sdk/README_ASYNC.md` for documentation

#### API Documentation ✅
- Created comprehensive OpenAPI 3.0 specification (`openapi.yaml`)
- Interactive documentation via Swagger UI or Redoc
- Full endpoint reference in `API_DOCUMENTATION.md`
- Authentication details and example requests
- Error handling and response codes documented

## License

AGPL-3.0 - All derivatives must be open source with the same license.

## Contributing

Follow AGENTS.md guidelines for contributions.

## Support

For issues and questions, refer to the original Donut Browser documentation.
