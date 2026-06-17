# API Documentation

This document provides comprehensive API documentation for Donut Browser Ultimate.

## OpenAPI/Swagger Specification

The complete OpenAPI 3.0 specification is available in [`openapi.yaml`](./openapi.yaml).

## Viewing the API Documentation

### Using Swagger UI

You can view interactive API documentation using Swagger UI:

```bash
# Install swagger-ui
npm install -g @apidevtools/swagger-cli

# Or use docker
docker run -p 8080:8080 -v $(pwd)/openapi.yaml:/openapi.yaml swaggerapi/swagger-ui
```

Then navigate to `http://localhost:8080` to view the interactive documentation.

### Using Redoc

```bash
# Install redoc-cli
npm install -g @redocly/cli

# Serve documentation
redocly serve-docs openapi.yaml
```

Navigate to the provided URL to view the documentation.

## API Endpoints

### Base URL

- Development: `http://localhost:10108`
- Production: `https://api.donutbrowser.com` (if available)

### Authentication

Admin endpoints require authentication via the `X-Admin-Token` header.

Set the token via environment variable:
```bash
export DONUT_ADMIN_TOKEN=your-secret-token
```

Default token: `admin-secret-token-change-me` (change in production!)

### Available Endpoints

#### Health & Status
- `GET /health` - Health check
- `GET /version` - Get API version

#### Profile Management
- `GET /profiles` - List all profiles
- `POST /profiles` - Create profile
- `GET /profiles/{id}` - Get profile details
- `PUT /profiles/{id}` - Update profile
- `DELETE /profiles/{id}` - Delete profile
- `POST /profiles/run` - Launch profile
- `POST /profiles/kill` - Stop profile

#### Workflow Management
- `GET /workflows` - List workflows
- `POST /workflows` - Create workflow
- `POST /workflows/{id}/execute` - Execute workflow

#### AI Automation
- `POST /ai/task` - Run AI task
- `GET /ai/task/{id}/status` - Get AI task status
- `POST /ai/task/{id}/stop` - Stop AI task

#### Batch Operations
- `POST /batch/profiles` - Create profiles in batch

#### Admin Operations (Authenticated)
- `GET /admin/profiles` - Get all profiles (admin)
- `DELETE /admin/profiles/{id}` - Delete profile (admin)
- `GET /admin/export` - Export all data (admin)
- `GET /admin/status` - Get system status (admin)

## Example Requests

### Create Profile

```bash
curl -X POST http://localhost:10108/profiles \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Profile",
    "browser": "wayfern",
    "tags": ["automation"]
  }'
```

### Launch Profile

```bash
curl -X POST http://localhost:10108/profiles/run \
  -H "Content-Type: application/json" \
  -d '{
    "profile_id": "profile-123",
    "url": "https://example.com",
    "headless": false
  }'
```

### Run AI Task

```bash
curl -X POST http://localhost:10108/ai/task \
  -H "Content-Type: application/json" \
  -d '{
    "profile_id": "profile-123",
    "task_definition": "Login to Gmail with email test@example.com",
    "vision_enabled": true,
    "max_steps": 100
  }'
```

### Batch Create Profiles

```bash
curl -X POST http://localhost:10108/batch/profiles \
  -H "Content-Type: application/json" \
  -d '{
    "profiles": [
      {"name": "Profile 1", "browser": "wayfern"},
      {"name": "Profile 2", "browser": "wayfern"},
      {"name": "Profile 3", "browser": "wayfern"}
    ],
    "concurrent_limit": 5
  }'
```

### Admin Get System Status

```bash
curl -X GET http://localhost:10108/admin/status \
  -H "X-Admin-Token: your-secret-token"
```

## Response Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 500 | Server Error |

## Error Response Format

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {
    "additional": "information"
  }
}
```

## Rate Limiting

Rate limits are applied to prevent abuse:
- Standard endpoints: 100 requests per minute
- Admin endpoints: 50 requests per minute
- Batch operations: 10 requests per minute

Rate limit headers are included in responses:
- `X-RateLimit-Limit`: Request limit
- `X-RateLimit-Remaining`: Remaining requests
- `X-RateLimit-Reset`: Reset timestamp

## SDK Integration

### Python SDK

See [`python-sdk/README.md`](./python-sdk/README.md) for sync SDK documentation.

See [`python-sdk/README_ASYNC.md`](./python-sdk/README_ASYNC.md) for async SDK documentation.

### MCP Integration

See the main [`README.md`](./README.md) for MCP server documentation.

## Testing the API

### Using cURL

```bash
# Health check
curl http://localhost:10108/health

# List profiles
curl http://localhost:10108/profiles

# Create profile
curl -X POST http://localhost:10108/profiles \
  -H "Content-Type: application/json" \
  -d '{"name": "Test", "browser": "wayfern"}'
```

### Using Python

```python
import requests

# Health check
response = requests.get("http://localhost:10108/health")
print(response.json())

# Create profile
response = requests.post(
    "http://localhost:10108/profiles",
    json={"name": "Test", "browser": "wayfern"}
)
print(response.json())
```

### Using JavaScript/Node.js

```javascript
// Health check
const response = await fetch('http://localhost:10108/health');
const data = await response.json();
console.log(data);

// Create profile
const response = await fetch('http://localhost:10108/profiles', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ name: 'Test', browser: 'wayfern' })
});
const data = await response.json();
console.log(data);
```

## WebSocket Support (Future)

Real-time updates via WebSocket are planned for future releases. This will enable:
- Live status updates for long-running tasks
- Real-time profile status changes
- Workflow execution progress streaming

## API Versioning

The API uses URL versioning:
- Current version: v1
- Base path: `/api/v1`

Backward compatibility is maintained for at least one major version.

## Support

For API-related issues and questions:
1. Check the OpenAPI specification in [`openapi.yaml`](./openapi.yaml)
2. Review the main [`README.md`](./README.md)
3. Check the original Donut Browser documentation
