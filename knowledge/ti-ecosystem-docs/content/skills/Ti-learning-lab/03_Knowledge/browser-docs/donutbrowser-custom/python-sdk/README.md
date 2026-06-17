# Donut Browser Python SDK

Simple Python SDK for Donut Browser automation with AI agent integration.

## Installation

```bash
pip install -r requirements.txt
```

## Quick Start

```python
from donut_browser import DonutBrowser

# Initialize SDK
browser = DonutBrowser()

# Create profile
profile = browser.create_profile(
    name="My Profile",
    browser="wayfern",
    tags=["automation"]
)

# Launch profile
browser.launch_profile(
    profile_id=profile["id"],
    url="https://example.com"
)

# Run AI task
result = browser.run_ai_task(
    profile_id=profile["id"],
    task_definition="Login to the website",
    vision_enabled=True
)

# Stop profile
browser.stop_profile(profile["id"])
```

## Features

- **Profile Management**: Create, list, get, delete, launch, stop profiles
- **AI Automation**: Run AI-powered tasks with natural language
- **Computer Vision**: Detect interactive elements on screenshots
- **Workflow Execution**: Create and execute automation workflows

## Examples

See `example_gmail_automation.py` for Gmail automation example.

## API Reference

### Profile Management

- `create_profile(name, browser, proxy_id, tags)` - Create new profile
- `list_profiles()` - List all profiles
- `get_profile(profile_id)` - Get profile details
- `delete_profile(profile_id)` - Delete profile
- `launch_profile(profile_id, url, headless)` - Launch profile
- `stop_profile(profile_id)` - Stop profile

### AI Automation

- `run_ai_task(profile_id, task_definition, vision_enabled, max_steps)` - Run AI task
- `get_agent_status(task_id)` - Get agent status
- `stop_agent(task_id)` - Stop agent

### Computer Vision

- `detect_elements(profile_id, filter_interactive)` - Detect elements

### Workflow Management

- `create_workflow(name, blocks, parameters)` - Create workflow
- `execute_workflow(workflow_id, profile_id)` - Execute workflow

## License

AGPL-3.0
