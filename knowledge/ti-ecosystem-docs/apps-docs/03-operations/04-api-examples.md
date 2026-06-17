# API Examples - Ví Dụ API

> **Project**: Ti Router  
> **Version**: 0.5.0  
> **Language**: Tiếng Việt  
> **Last Updated**: 2026-04-30

---

## Tổng Quan

Tài liệu này cung cấp các ví dụ chi tiết về cách sử dụng Ti Router API. Bao gồm:
- Chat Completions API
- A/B Testing API
- Canary Deployment API
- Feature Flags API
- Blue-Green Deployment API
- Analytics API

---

## Base URL

```
http://localhost:20128
```

---

## Authentication

Tất cả API endpoints yêu cầu Bearer token (API key):

```bash
curl -H "Authorization: Bearer YOUR_API_KEY" http://localhost:20128/api/v1/chat/completions
```

---

## Chat Completions API

### Basic Chat Completion

```bash
curl -X POST http://localhost:20128/api/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "gpt-4",
    "messages": [
      {
        "role": "user",
        "content": "Hello, how are you?"
      }
    ]
  }'
```

**Response:**
```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "created": 1677652288,
  "model": "gpt-4",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "I'm doing great, thank you for asking!"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 20,
    "total_tokens": 30
  }
}
```

### Streaming Chat Completion

```bash
curl -X POST http://localhost:20128/api/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "gpt-4",
    "messages": [
      {
        "role": "user",
        "content": "Tell me a story"
      }
    ],
    "stream": true
  }'
```

**Response (SSE stream):**
```
data: {"id":"chatcmpl-123","choices":[{"delta":{"content":"Once"}}]}
data: {"id":"chatcmpl-123","choices":[{"delta":{"content":" upon"}}]}
data: {"id":"chatcmpl-123","choices":[{"delta":{"content":" a"}}]}
...
data: [DONE]
```

### Provider-Specific Request

```bash
curl -X POST http://localhost:20128/api/v1/providers/openai/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "gpt-4",
    "messages": [
      {
        "role": "user",
        "content": "Hello"
      }
    ]
  }'
```

---

## A/B Testing API

### Create A/B Test

```bash
curl -X POST http://localhost:20128/api/ab-tests \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "name": "Test Routing Strategies",
    "description": "Compare RoundRobin vs LeastLatency",
    "variants": [
      {
        "id": "variant_a",
        "name": "RoundRobin Strategy",
        "weight": 0.5,
        "config": {
          "strategy": "round_robin"
        }
      },
      {
        "id": "variant_b",
        "name": "LeastLatency Strategy",
        "weight": 0.5,
        "config": {
          "strategy": "least_latency"
        }
      }
    ],
    "traffic_allocation": {
      "strategy": "user_hash"
    }
  }'
```

**Response:**
```json
{
  "id": "ab_test_1234567890",
  "name": "Test Routing Strategies",
  "status": "draft",
  "created_at": "2026-04-30T00:00:00Z"
}
```

### Start A/B Test

```bash
curl -X POST http://localhost:20128/api/ab-tests/ab_test_1234567890 \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### Get Test Details

```bash
curl -X GET http://localhost:20128/api/ab-tests/ab_test_1234567890 \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Response:**
```json
{
  "id": "ab_test_1234567890",
  "name": "Test Routing Strategies",
  "status": "running",
  "metrics": {
    "variant_metrics": {
      "variant_a": {
        "requests": 1000,
        "successes": 950,
        "errors": 50,
        "avg_latency_ms": 150.5
      },
      "variant_b": {
        "requests": 1000,
        "successes": 980,
        "errors": 20,
        "avg_latency_ms": 120.3
      }
    }
  }
}
```

### Get Winner

```bash
curl -X GET http://localhost:20128/api/ab-tests/ab_test_1234567890/winner \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Response:**
```json
{
  "id": "variant_b",
  "name": "LeastLatency Strategy",
  "reason": "Higher success rate (98% vs 95%)"
}
```

### Stop A/B Test

```bash
curl -X DELETE http://localhost:20128/api/ab-tests/ab_test_1234567890 \
  -H "Authorization: Bearer YOUR_API_KEY"
```

---

## Canary Deployment API

### Create Canary Deployment

```bash
curl -X POST http://localhost:20128/api/canary/deployments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "name": "New Model Version",
    "description": "Gradual rollout of gpt-5",
    "versions": [
      {
        "id": "v1",
        "name": "Version 1.0 (gpt-4)",
        "weight": 0.9,
        "is_stable": true,
        "config": {
          "model": "gpt-4"
        }
      },
      {
        "id": "v2",
        "name": "Version 2.0 (gpt-5)",
        "weight": 0.1,
        "is_stable": false,
        "config": {
          "model": "gpt-5"
        }
      }
    ],
    "strategy": {
      "type": "gradual",
      "increment": 0.1,
      "interval": "5m",
      "success_threshold": 0.95,
      "error_threshold": 0.05,
      "auto_promote": true,
      "auto_rollback": true
    }
  }'
```

**Response:**
```json
{
  "id": "canary_1234567890",
  "name": "New Model Version",
  "status": "pending",
  "created_at": "2026-04-30T00:00:00Z"
}
```

### Start Canary Deployment

```bash
curl -X POST http://localhost:20128/api/canary/deployments/canary_1234567890 \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### Promote Version

```bash
curl -X PATCH http://localhost:20128/api/canary/deployments/canary_1234567890 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "version_id": "v2",
    "increment": 0.1
  }'
```

### Rollback Deployment

```bash
curl -X DELETE http://localhost:20128/api/canary/deployments/canary_1234567890 \
  -H "Authorization: Bearer YOUR_API_KEY"
```

---

## Feature Flags API

### Create Feature Flag

```bash
curl -X POST http://localhost:20128/api/feature-flags \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "name": "enable_new_api",
    "description": "Enable new API for beta users",
    "type": "percentage",
    "enabled": true
  }'
```

**Response:**
```json
{
  "id": "flag_1234567890",
  "name": "enable_new_api",
  "type": "percentage",
  "enabled": true
}
```

### Add Rule to Feature Flag

```bash
curl -X POST http://localhost:20128/api/feature-flags/flag_1234567890/rules \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "type": "percentage",
    "value": 10,
    "condition": "less_than"
  }'
```

### Check if Flag is Enabled

```bash
curl -X POST http://localhost:20128/api/feature-flags/flag_1234567890/check \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "user_id": "user123"
  }'
```

**Response:**
```json
{
  "enabled": true,
  "variant": "variant_a"
}
```

### Update Feature Flag

```bash
curl -X PATCH http://localhost:20128/api/feature-flags/flag_1234567890 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "enabled": false
  }'
```

---

## Blue-Green Deployment API

### Create Blue-Green Deployment

```bash
curl -X POST http://localhost:20128/api/blue-green/deployments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "name": "Infrastructure Upgrade",
    "description": "Zero-downtime infrastructure upgrade",
    "blue": {
      "id": "env_blue",
      "name": "Blue Environment",
      "version": "1.0.0",
      "config": {
        "provider": "openai"
      },
      "is_healthy": true
    },
    "green": {
      "id": "env_green",
      "name": "Green Environment",
      "version": "2.0.0",
      "config": {
        "provider": "anthropic"
      },
      "is_healthy": true
    }
  }'
```

**Response:**
```json
{
  "id": "bg_1234567890",
  "name": "Infrastructure Upgrade",
  "current": "blue",
  "status": "pending"
}
```

### Start Deployment

```bash
curl -X POST http://localhost:20128/api/blue-green/deployments/bg_1234567890 \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### Switch Environment

```bash
curl -X PATCH http://localhost:20128/api/blue-green/deployments/bg_1234567890 \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### Rollback

```bash
curl -X DELETE http://localhost:20128/api/blue-green/deployments/bg_1234567890 \
  -H "Authorization: Bearer YOUR_API_KEY"
```

---

## Analytics API

### Get Analytics Metrics

```bash
curl -X GET http://localhost:20128/api/analytics/metrics \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Response:**
```json
{
  "requests": {
    "total": 100000,
    "success": 98000,
    "errors": 2000,
    "rps": 50.5,
    "rpm": 3000
  },
  "latency": {
    "avg": 150.5,
    "p50": 120.3,
    "p95": 200.7,
    "p99": 350.2
  },
  "tokens": {
    "prompt": 5000000,
    "completion": 10000000,
    "total": 15000000
  },
  "cost": {
    "total_usd": 150.5,
    "per_provider": {
      "openai": 100.0,
      "anthropic": 50.5
    }
  }
}
```

### Get Health Scores

```bash
curl -X GET http://localhost:20128/api/analytics/health \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Response:**
```json
{
  "providers": {
    "openai": 0.95,
    "anthropic": 0.92,
    "google": 0.88
  },
  "routes": {
    "default": 0.90
  }
}
```

### Get Percentile Metrics

```bash
curl -X GET "http://localhost:20128/api/analytics/percentiles?provider=openai" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Response:**
```json
{
  "latency_ms": {
    "p50": 120.3,
    "p95": 200.7,
    "p99": 350.2
  }
}
```

---

## Error Handling

### Common Error Codes

- **400 Bad Request**: Invalid request body
- **401 Unauthorized**: Invalid or missing API key
- **429 Too Many Requests**: Rate limit exceeded
- **502 Bad Gateway**: All upstream providers failed
- **503 Service Unavailable**: Service temporarily unavailable

### Error Response Format

```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Please retry later.",
    "retry_after": 60
  }
}
```

---

## Best Practices

### 1. Use Streaming cho Long Responses
```bash
curl -X POST http://localhost:20128/api/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "gpt-4",
    "messages": [...],
    "stream": true
  }'
```

### 2. Set Appropriate Timeouts
```bash
curl -X POST http://localhost:20128/api/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  --max-time 120 \
  -d '{...}'
```

### 3. Handle Rate Limits
```bash
response=$(curl -X POST http://localhost:20128/api/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{...}')

retry_after=$(echo "$response" | jq -r '.error.retry_after')
if [ "$retry_after" != "null" ]; then
  sleep $retry_after
  # Retry request
fi
```

### 4. Monitor Usage
```bash
curl -X GET http://localhost:20128/api/analytics/metrics \
  -H "Authorization: Bearer YOUR_API_KEY"
```

---

## Python Examples

### Basic Chat Completion

```python
import requests

url = "http://localhost:20128/api/v1/chat/completions"
headers = {
    "Content-Type": "application/json",
    "Authorization": "Bearer YOUR_API_KEY"
}
data = {
    "model": "gpt-4",
    "messages": [
        {"role": "user", "content": "Hello, how are you?"}
    ]
}

response = requests.post(url, headers=headers, json=data)
print(response.json())
```

### Streaming Chat Completion

```python
import requests

url = "http://localhost:20128/api/v1/chat/completions"
headers = {
    "Content-Type": "application/json",
    "Authorization": "Bearer YOUR_API_KEY"
}
data = {
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Tell me a story"}],
    "stream": True
}

response = requests.post(url, headers=headers, json=data, stream=True)
for line in response.iter_lines():
    if line:
        print(line.decode('utf-8'))
```

### A/B Testing

```python
import requests

# Create test
url = "http://localhost:20128/api/ab-tests"
data = {
    "name": "Test Routing Strategies",
    "variants": [
        {"id": "variant_a", "name": "RoundRobin", "weight": 0.5},
        {"id": "variant_b", "name": "LeastLatency", "weight": 0.5}
    ],
    "traffic_allocation": {"strategy": "user_hash"}
}
response = requests.post(url, headers=headers, json=data)
test_id = response.json()["id"]

# Start test
requests.post(f"http://localhost:20128/api/ab-tests/{test_id}", headers=headers)

# Get winner
winner = requests.get(f"http://localhost:20128/api/ab-tests/{test_id}/winner", headers=headers)
print(winner.json())
```

---

## JavaScript/Node.js Examples

### Basic Chat Completion

```javascript
const fetch = require('node-fetch');

const url = 'http://localhost:20128/api/v1/chat/completions';
const headers = {
  'Content-Type': 'application/json',
  'Authorization': 'Bearer YOUR_API_KEY'
};
const data = {
  model: 'gpt-4',
  messages: [
    { role: 'user', content: 'Hello, how are you?' }
  ]
};

fetch(url, {
  method: 'POST',
  headers: headers,
  body: JSON.stringify(data)
})
  .then(response => response.json())
  .then(result => console.log(result));
```

### Streaming Chat Completion

```javascript
const fetch = require('node-fetch');

const url = 'http://localhost:20128/api/v1/chat/completions';
const headers = {
  'Content-Type': 'application/json',
  'Authorization': 'Bearer YOUR_API_KEY'
};
const data = {
  model: 'gpt-4',
  messages: [{ role: 'user', content: 'Tell me a story' }],
  stream: true
};

fetch(url, {
  method: 'POST',
  headers: headers,
  body: JSON.stringify(data)
})
  .then(response => {
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    
    function read() {
      reader.read().then(({ done, value }) => {
        if (done) return;
        console.log(decoder.decode(value));
        read();
      });
    }
    read();
  });
```

---

## Testing with cURL

### Test Chat Completion

```bash
# Test basic chat completion
curl -X POST http://localhost:20128/api/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{"model":"gpt-4","messages":[{"role":"user","content":"Hello"}]}'

# Test streaming
curl -X POST http://localhost:20128/api/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{"model":"gpt-4","messages":[{"role":"user","content":"Hello"}],"stream":true}'
```

### Test A/B Testing

```bash
# Create test
curl -X POST http://localhost:20128/api/ab-tests \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{"name":"Test","variants":[{"id":"a","weight":0.5},{"id":"b","weight":0.5}],"traffic_allocation":{"strategy":"user_hash"}}'

# List tests
curl -X GET http://localhost:20128/api/ab-tests \
  -H "Authorization: Bearer YOUR_API_KEY"
```

---

## Next Steps

1. Implement actual E2E tests với running server
2. Add more language examples (Ruby, Go, Java)
3. Add WebSocket examples cho streaming
4. Add batch request examples
5. Add error handling examples
