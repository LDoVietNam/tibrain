# Task Agent - Autonomous Task Management

## Overview

The Task Agent is an autonomous agent responsible for automatically triaging and managing tasks in the Spectre AI system. It periodically scans the task board, applies triage rules to tasks, and updates them with standardized information.

## Architecture

The agent follows a scheduled pattern:
1. **Periodic Execution**: Runs on a configurable interval (default 60 seconds)
2. **Task Listing**: Fetches tasks from the task board via MCP Gateway
3. **Triage Processing**: Applies standardized triage rules to tasks
4. **Task Updates**: Updates tasks with normalized information
5. **Loop Continuation**: Repeats on next interval

## Key Features

- **MCP Gateway Integration**: Communicates exclusively through the Spectre MCP Gateway
- **Idempotent Operations**: Safe to run multiple times without duplicate effects
- **Configurable**: Highly customizable through environment variables
- **Automatic**: Runs continuously without manual intervention

## Configuration

Configuration is managed through environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `TASK_AGENT_ENABLED` | `1` | Enable/disable the agent |
| `TASK_AGENT_INTERVAL_SECONDS` | `60` | Polling interval in seconds |
| `TASK_AGENT_MAX_PER_TICK` | `25` | Maximum tasks processed per cycle |
| `TASK_AGENT_TRIAGED_TAG` | `triaged:true` | Tag applied to triaged tasks |
| `TASK_AGENT_AGENT_TAG` | `agent:task` | Tag identifying agent-managed tasks |
| `TASK_AGENT_DEFAULT_PRIORITY` | `P2` | Default priority for new tasks |
| `TASK_AGENT_DEFAULT_LANE` | `other` | Default lane for new tasks |
| `TASK_AGENT_DEFAULT_POLICY` | `balanced` | Default policy for new tasks |

## How It Works

1. **Initialization**: Agent starts with configured settings
2. **Tick Cycle**: 
   - Lists tasks from the task board (with buffer limit)
   - Processes each task for triage
   - Applies updates if triage rules are met
   - Limits updates to `max_per_tick` per cycle
3. **Continuous Operation**: Runs in a loop until stopped

## Testing

### Unit Tests

The agent has comprehensive unit tests that verify:
- Agent initialization
- Tick logic with mocked gateway interactions
- Proper task processing and update behavior

To run unit tests:
```bash
py -m pytest tests/test_task_agent_unit.py -v
```

### Integration Tests

Integration tests verify:
- Configuration correctness
- Core logic flow
- Task structure validation

To run integration tests:
```bash
py -m pytest tests/test_task_agent_integration.py -v
```

## Requirements

- Python 3.10+
- HTTP access to MCP Gateway (default: `http://127.0.0.1:8000`)
- Properly configured environment variables

## Running the Agent

To run the agent:
```bash
py -m agents.task_agent.task_agent
```

Note: The agent requires the MCP Gateway to be running at the configured URL. Without it, the agent will log connection errors but will continue running.

## Implementation Details

The agent is implemented in `task_agent.py` and uses:
- `GatewayClient` for MCP Gateway communication
- `TaskAgentConfig` for configuration management
- `extract_tasks` and `build_triage_update` for task processing
- Async I/O for efficient network operations

## Security Considerations

- All communications happen through the MCP Gateway
- No direct database or backend API access
- Authentication handled via API keys in environment variables
- Idempotent operations prevent duplicate processing

## Monitoring

The agent logs:
- Start/stop events
- Tick completion statistics
- Connection errors
- Task processing results

## Troubleshooting

### Common Issues

1. **Connection Errors**: Gateway not running at configured URL
2. **Permission Denied**: Invalid API key or insufficient permissions
3. **Rate Limiting**: Exceeding maximum tasks per tick

### Solutions

1. Ensure MCP Gateway is running
2. Verify `GATEWAY_API_KEY` environment variable
3. Adjust `TASK_AGENT_MAX_PER_TICK` if needed
