# Research: Go SDK Patterns cho AI Agents

> Ngày: 2026-04-29 | Agent: claude

## 1. go-anthropic (liushuangls)

```go
client := anthropic.NewClient("api-key")
resp, err := client.CreateMessages(ctx, anthropic.MessagesRequest{...})

// Error handling
var e *anthropic.APIError
if errors.As(err, &e) { ... }
```

**Lessons**: Simple constructor, typed errors, callback streaming.

## 2. agent-sdk-go (Ingenimax)

```go
agent.NewAgent(
    agent.WithLLM(openaiClient),
    agent.WithMemory(memory.NewConversationBuffer()),
    agent.WithTools(tools...),
)
```

**Lessons**: Option pattern, modular (llm/memory/tools/logging).

## 3. REST API Client Best Practices

- Internal `http.Client` có thể override
- `do(method, endpoint, body, params)` là core
- JSON auto marshal/unmarshal
- Pointer fields cho optional values
- Default timeout (120s cho Devin)

## 4. So sánh Python vs Go cho Devin

| Feature | Python | Go |
|---------|--------|-----|
| Client | ClientProxy (V1/V3) | Client + WithAPIVersion() |
| Config | Class properties | Struct methods |
| HTTP | httpx | net/http |
| Poll | time.sleep() | time.Ticker + ctx.Done() |
| Error | Custom exception | error interface |
| JSON | json.load/dump | encoding/json |
