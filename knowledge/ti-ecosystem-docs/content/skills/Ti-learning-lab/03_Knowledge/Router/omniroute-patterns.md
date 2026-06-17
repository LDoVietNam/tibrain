# OmniRoute Architecture & Patterns

> **Source**: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\07_Repositories\router\OmniRoute\`
> **Applied to**: Ti Router (`Z:\10_WORKPLACE\Ti\apps\router\`)
> **Date**: 2026-05-04

## Overview

OmniRoute là một unified AI proxy/router với 160+ providers, multi-tier fallback, advanced routing strategies, MCP Server (29 tools), A2A Protocol, và Electron desktop app. Bài viết này ghi lại các architecture và patterns học được từ OmniRoute.

---

## Tech Stack

| Component | Technology |
|-----------|------------|
| **Runtime** | Next.js 16 (App Router), Node.js ≥18 <24, ES Modules |
| **Language** | TypeScript 5.9 (`src/`) + JavaScript (`open-sse/`, `electron/`) |
| **Database** | better-sqlite3 (SQLite) - WAL journaling |
| **Streaming** | SSE via `open-sse` internal workspace package |
| **Styling** | Tailwind CSS v4 |
| **i18n** | next-intl với 40+ languages |
| **Desktop** | Electron (cross-platform: Windows, macOS, Linux) |
| **Schemas** | Zod v4 cho tất cả API / MCP input validation |

---

## Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│                    Client Layer                             │
│  (Claude Code, Codex, Gemini CLI, Cursor, Cline, etc.)   │
└────────────────────┬────────────────────────────────────────┘
                     │ http://localhost:20128/v1
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                   API Route Layer                           │
│  (Next.js App Router - src/app/api/v1/)                   │
│  - CORS preflight                                            │
│  - Body validation (Zod)                                     │
│  - Optional auth (extractApiKey/isValidApiKey)               │
│  - API key policy enforcement                                │
│  - Handler delegation (open-sse)                             │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│              Request Pipeline (open-sse/)                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  Handlers     │  │  Executors    │  │  Translators  │     │
│  │  chatCore.ts  │→ │  base.ts      │→ │  translator/  │     │
│  │  embeddings.ts│  │  default.ts   │  │  request/     │     │
│  │  images.ts    │  │  cursor.ts    │  │  response/    │     │
│  │  ...          │  │  codex.ts     │  │               │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  Services     │  │  Transformer  │  │  MCP Server   │     │
│  │  combo.ts     │  │  responses... │  │  29 tools     │     │
│  │  rateLimit... │  │               │  │  3 transports │     │
│  │  tokenRefresh │  │               │  │  10 scopes    │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│               Data Layer (src/lib/db/)                      │
│  - Domain modules (22 files)                                 │
│  - Schema migrations (21 files)                              │
│  - Encryption at rest                                        │
│  - WAL journaling                                            │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│           Domain/Policy Layer (src/domain/)                  │
│  - Policy engine                                             │
│  - Cost rules                                                │
│  - Fallback policy                                          │
│  - Lockout policy                                            │
│  - Model availability                                       │
└─────────────────────────────────────────────────────────────┘
```

---

## 1. Database Architecture Pattern

### Location: `src/lib/db/`

### Pattern Components

#### 1.1 Core Infrastructure
```typescript
// core.ts - Singleton DB instance
import Database from "better-sqlite3";

export function getDbInstance(): Database.Database {
  // Returns singleton with WAL journaling
  // SCHEMA_SQL defines 15 base tables
  // Helpers: rowToCamel(), encryptConnectionFields()
}
```

**Key Points**:
- Singleton pattern cho DB instance
- WAL journaling cho concurrent reads
- Snake_case → camelCase conversion
- Encryption helpers cho sensitive fields

#### 1.2 Migration Pattern
```typescript
// migrationRunner.ts - Versioned migrations
import { runMigrations } from "./migrationRunner";

// Applies versioned SQL files from db/migrations/
// Tracks applied migrations in _omniroute_migrations table
// Each migration runs in a transaction
// Idempotent - safe to re-run
```

**Migrations Structure**:
```
db/migrations/
├── 001_initial_schema.sql
├── 002_add_provider_limits.sql
├── ...
└── 021_combo_call_log_targets.sql
```

**Key Points**:
- Versioned SQL files (001-021)
- Each migration = single responsibility
- Runs in transaction (atomic)
- Idempotent (safe to re-run)
- Tracks applied migrations

#### 1.3 Domain Module Pattern
```typescript
// providers.ts - Example domain module
import { getDbInstance } from "./core";

export function getProviderConnections(): ProviderConnection[] {
  const db = getDbInstance();
  const stmt = db.prepare("SELECT * FROM provider_connections");
  return stmt.all().map(rowToCamel);
}

export function createProviderConnection(conn: ProviderConnection): void {
  const db = getDbInstance();
  const stmt = db.prepare(`
    INSERT INTO provider_connections (id, provider, ...)
    VALUES (?, ?, ...)
  `);
  stmt.run(...);
}
```

**Key Points**:
- Each module owns specific table(s)
- Import `getDbInstance()` from `core.ts`
- Export CRUD functions
- No raw SQL in routes
- `localDb.ts` is re-export layer only (no logic)

#### 1.4 Domain Modules List (22 total)

| Module | Tables | Responsibility |
|--------|--------|---------------|
| `providers.ts` | `provider_connections` | OAuth/API key provider registration |
| `models.ts` | `models` | Model definitions, capabilities, pricing |
| `combos.ts` | `combos`, `combo_targets` | Combo routing configs |
| `apiKeys.ts` | `api_keys` | API key lifecycle, scopes, quota |
| `settings.ts` | `settings` | KV store for system config |
| `backup.ts` | - | Backup export/import ops |
| `proxies.ts` | `proxies` | MITM proxy configs |
| `prompts.ts` | `prompts` | Reusable prompt templates |
| `webhooks.ts` | `webhooks` | Event-driven webhooks |
| `detailedLogs.ts` | `detailed_logs` | Per-request audit logging |
| `quotaSnapshots.ts` | `quota_snapshots` | Historical quota usage |
| `modelComboMappings.ts` | `model_combo_mappings` | Map models to combo defaults |
| `cliToolState.ts` | `cli_tool_state` | CLI-specific persistent state |
| `encryption.ts` | - | Encryption/decryption helpers |
| `readCache.ts` | - | In-memory cache for read-heavy ops |
| `secrets.ts` | `secrets` | Encrypted secret storage |
| `stateReset.ts` | - | Wipe/reset DB state |
| `contextHandoffs.ts` | `context_handoffs` | Session context for agent handoff |
| `migrations/` | - | Versioned SQL schema evolution |
| `core.ts` | - | Singleton DB instance, helpers |

---

## 2. Executor Pattern

### Location: `open-sse/executors/`

### Pattern Components

#### 2.1 BaseExecutor (Strategy Pattern)
```typescript
// base.ts - Abstract base class
export class BaseExecutor {
  provider: string;
  config: ProviderConfig;

  constructor(provider: string, config: ProviderConfig) {
    this.provider = provider;
    this.config = config;
  }

  // Override in subclass for provider-specific behavior
  buildUrl(model: string, stream: boolean, urlIndex: number): string {
    // Default implementation
    const baseUrls = this.getBaseUrls();
    return baseUrls[urlIndex] || baseUrls[0];
  }

  buildHeaders(credentials: ProviderCredentials, stream: boolean): Record<string, string> {
    const headers = {
      "Content-Type": "application/json",
      ...this.config.headers,
    };

    if (credentials.accessToken) {
      headers["Authorization"] = `Bearer ${credentials.accessToken}`;
    } else if (credentials.apiKey) {
      headers["Authorization"] = `Bearer ${credentials.apiKey}`;
    }

    headers["Accept"] = stream ? "text/event-stream" : "application/json";
    return headers;
  }

  transformRequest(model: string, body: unknown, stream: boolean): unknown {
    // Override in subclass for provider-specific transformations
    return body;
  }

  async execute({ model, body, stream, credentials, signal }: ExecuteInput) {
    // Main execution logic with fallback and retry
    for (let urlIndex = 0; urlIndex < fallbackCount; urlIndex++) {
      const url = this.buildUrl(model, stream, urlIndex);
      const headers = this.buildHeaders(credentials, stream);
      const transformedBody = this.transformRequest(model, body, stream);

      try {
        const response = await fetch(url, {
          method: "POST",
          headers,
          body: JSON.stringify(transformedBody),
          signal,
        });

        if (!response.ok && this.shouldRetry(response.status, urlIndex)) {
          continue; // Try next URL
        }

        return response;
      } catch (error) {
        // Handle error, try next URL
      }
    }
  }

  shouldRetry(status: number, urlIndex: number): boolean {
    return status === 429 && urlIndex + 1 < this.getFallbackCount();
  }

  async refreshCredentials(credentials: ProviderCredentials): Promise<ProviderCredentials | null> {
    // Override in subclass for OAuth token refresh
    return null;
  }
}
```

**Key Points**:
- Strategy pattern - subclasses override specific methods
- Fallback mechanism - multiple base URLs
- Retry logic with exponential backoff
- Credential refresh - auto refresh when expired
- Timeout handling with abort signals

#### 2.2 Provider-Specific Executors
```typescript
// cursor.ts - Cursor-specific executor
export class CursorExecutor extends BaseExecutor {
  buildUrl(model: string, stream: boolean): string {
    // Cursor-specific URL building
    return `https://api2.cursor.sh/v1/chat/completions`;
  }

  buildHeaders(credentials: ProviderCredentials): Record<string, string> {
    const headers = super.buildHeaders(credentials);
    // Add Cursor-specific headers
    headers["x-cursor-client"] = "cli";
    return headers;
  }

  transformRequest(model: string, body: unknown): unknown {
    // Cursor-specific request transformation
    return body;
  }
}

// codex.ts - Codex-specific executor
export class CodexExecutor extends BaseExecutor {
  buildUrl(model: string, stream: boolean): string {
    return `https://api.githubcopilot.com/copilot_internal/v2/chat/completions`;
  }

  async refreshCredentials(credentials: ProviderCredentials): Promise<ProviderCredentials | null> {
    // OAuth token refresh for Codex
    const refreshed = await refreshCodexToken(credentials.refreshToken);
    return refreshed;
  }
}
```

**Key Points**:
- Extend BaseExecutor
- Override only what differs (URL, headers, transform)
- Provider-specific OAuth refresh
- Keep common logic in base class

#### 2.3 Executor Factory
```typescript
// index.ts - Factory pattern
export function getExecutor(provider: string, config: ProviderConfig): BaseExecutor {
  switch (provider) {
    case "cursor":
      return new CursorExecutor(provider, config);
    case "codex":
      return new CodexExecutor(provider, config);
    case "claude":
      return new ClaudeExecutor(provider, config);
    default:
      return new DefaultExecutor(provider, config);
  }
}
```

**Key Points**:
- Factory pattern for executor creation
- Default executor for most OpenAI-compatible providers
- Provider-specific executors for special cases

---

## 3. Combo Routing Engine Pattern

### Location: `open-sse/services/combo.ts`

### Pattern Components

#### 3.1 Combo Resolution
```typescript
// combo.ts - Combo routing engine
type ResolvedComboTarget = {
  kind: "model";
  stepId: string;
  executionKey: string;
  modelStr: string;
  provider: string;
  providerId: string | null;
  connectionId: string | null;
  weight: number;
  label: string | null;
};

async function resolveComboTargets(
  comboConfig: ComboConfig,
  requestContext: RequestContext
): Promise<ResolvedComboTarget[]> {
  // Expand combo config into ordered targets
  const targets: ResolvedComboTarget[] = [];

  for (const step of comboConfig.steps) {
    const provider = step.provider;
    const model = step.model;
    const connectionId = selectConnection(provider, step.strategy);

    targets.push({
      kind: "model",
      stepId: step.id,
      executionKey: `${provider}:${model}:${connectionId}`,
      modelStr: model,
      provider,
      providerId: step.providerId,
      connectionId,
      weight: step.weight || 1,
      label: step.label,
    });
  }

  // Apply strategy ordering
  return applyStrategy(targets, comboConfig.strategy);
}
```

**Key Points**:
- Expand combo config into ordered targets
- Each target = provider + model + account + credentials
- Strategy determines target ordering

#### 3.2 Routing Strategies (13 total)
```typescript
// Strategies
type Strategy =
  | "priority"      // Ordered list
  | "weighted"      // Probabilistic
  | "fill-first"    // Fill quota first
  | "round-robin"   // Round-robin
  | "P2C"           // Power of two choices
  | "random"        // Random selection
  | "least-used"    // Least used recently
  | "cost-optimized" // Cheapest first
  | "strict-random" // Random with constraints
  | "auto"          // AI-powered auto routing
  | "lkgp"          // Last known good provider
  | "context-optimized" // Context-aware routing
  | "context-relay"; // Relay context between targets

function applyStrategy(targets: ResolvedComboTarget[], strategy: Strategy): ResolvedComboTarget[] {
  switch (strategy) {
    case "priority":
      return targets; // Already ordered
    case "weighted":
      return weightedShuffle(targets);
    case "round-robin":
      return roundRobin(targets);
    case "cost-optimized":
      return sortByCost(targets);
    // ... other strategies
  }
}
```

**Key Points**:
- 13 routing strategies
- Strategy pattern - pluggable routing logic
- Cost-aware, latency-aware, health-aware routing

#### 3.3 Combo Execution
```typescript
async function handleComboChat(
  request: ChatRequest,
  comboConfig: ComboConfig
): Promise<ChatResponse> {
  const targets = await resolveComboTargets(comboConfig, request);
  let lastError: Error | null = null;

  for (const target of targets) {
    try {
      // Execute request for this target
      const response = await handleSingleModel(request, target);

      // Success - return response
      return response;
    } catch (error) {
      lastError = error;

      // Check if should fallback
      if (shouldFallback(error, target)) {
        continue; // Try next target
      }

      // Circuit breaker check
      if (isProviderCircuitBreakerOpen(target.provider)) {
        continue; // Skip this provider
      }
    }
  }

  // All targets failed
  throw new Error(`All combo targets failed: ${lastError?.message}`);
}
```

**Key Points**:
- Iterate through targets in order
- Fallback on failure
- Circuit breaker integration
- Return first successful response

---

## 4. Translator Pattern

### Location: `open-sse/translator/`

### Pattern Components

#### 4.1 Format Registry
```typescript
// formats.ts - Format constants
export const FORMATS = {
  OPENAI: "openai",
  ANTHROPIC: "anthropic",
  GEMINI: "gemini",
  RESPONSES: "responses",
} as const;

export type Format = (typeof FORMATS)[keyof typeof FORMATS];
```

#### 4.2 Request Translation
```typescript
// index.ts - Request translation
export function translateRequest(
  sourceFormat: string,
  targetFormat: string,
  model: string,
  body: unknown,
  stream: boolean,
  credentials: ProviderCredentials | null,
  provider: string | null,
  options?: {
    normalizeToolCallId?: boolean;
    preserveDeveloperRole?: boolean;
    preserveCacheControl?: boolean;
  }
): unknown {
  let result = body;

  // Normalize to OpenAI format first (if needed)
  if (sourceFormat !== FORMATS.OPENAI) {
    result = normalizeToOpenAI(result, sourceFormat);
  }

  // Apply tool call normalization
  if (options?.normalizeToolCallId) {
    result = normalizeToolCallIds(result);
  }

  // Apply role normalization
  result = normalizeRoles(result, options?.preserveDeveloperRole);

  // Translate to target format
  if (targetFormat !== FORMATS.OPENAI) {
    result = translateFromOpenAI(result, targetFormat, model, provider);
  }

  return result;
}
```

**Key Points**:
- Normalize to OpenAI format first (canonical format)
- Apply transformations (tool calls, roles, etc.)
- Translate to target format
- Preserve cache control for Claude Code

#### 4.3 Response Translation
```typescript
// response/index.ts - Response translation
export function translateResponse(
  sourceFormat: string,
  targetFormat: string,
  response: unknown,
  model: string
): unknown {
  // Reverse of request translation
  if (sourceFormat === targetFormat) {
    return response;
  }

  // Convert source format to OpenAI
  const openaiResponse = normalizeToOpenAI(response, sourceFormat);

  // Convert OpenAI to target format
  return translateFromOpenAI(openaiResponse, targetFormat, model);
}
```

**Key Points**:
- Reverse of request translation
- OpenAI as canonical format
- Handle streaming responses

---

## 5. MCP Server Pattern

### Location: `open-sse/mcp-server/`

### Pattern Components

#### 5.1 MCP Server Architecture
```
AI Agent (Claude Desktop, Cursor, VS Code)
  ↓ MCP Protocol (stdio or HTTP)
OmniRoute MCP Server
  ├── Scope Enforcement (10 scopes)
  ├── 29 Tools (Phase 1: 8 essential, Phase 2: 21 advanced)
  └── Audit Logger (SHA-256/SQLite)
  ↓ HTTP (internal)
OmniRoute Gateway (port 20128)
```

#### 5.2 Tool Definition Pattern
```typescript
// tools/getHealth.ts - Example tool
import { z } from "zod";

const inputSchema = z.object({
  detailed: z.boolean().optional().default(false),
});

export const getHealthTool = {
  name: "omniroute_get_health",
  description: "Get gateway health status",
  inputSchema,
  handler: async (input: z.infer<typeof inputSchema>) => {
    const health = await getGatewayHealth(input.detailed);
    return {
      content: [{ type: "text", text: JSON.stringify(health, null, 2) }],
    };
  },
};
```

**Key Points**:
- Zod schema cho input validation
- Async handler function
- Return MCP tool response format
- Type-safe với TypeScript

#### 5.3 Scope Enforcement
```typescript
// scopeEnforcement.ts - Scope-based access control
type Scope =
  | "read:health"
  | "read:combos"
  | "read:quota"
  | "read:usage"
  | "read:models"
  | "execute:completions"
  | "write:combos"
  | "write:budget"
  | "write:resilience"
  | "*";

const TOOL_SCOPES: Record<string, Scope[]> = {
  "omniroute_get_health": ["read:health"],
  "omniroute_list_combos": ["read:combos"],
  "omniroute_switch_combo": ["write:combos"],
  "omniroute_route_request": ["execute:completions"],
  // ...
};

export function checkScope(toolName: string, userScopes: Scope[]): boolean {
  const requiredScopes = TOOL_SCOPES[toolName];
  if (!requiredScopes) return false;

  // Check if user has any required scope
  return requiredScopes.some(scope =>
    userScopes.includes(scope) || userScopes.includes("*")
  );
}
```

**Key Points**:
- 10 fine-grained scopes
- Tool-to-scope mapping
- Wildcard support (`*`)
- Enforced before tool execution

#### 5.4 Audit Logging
```typescript
// audit.ts - Tool call audit logging
export async function logToolCall(
  toolName: string,
  input: unknown,
  output: unknown,
  durationMs: number,
  apiKeyId: string | null
): Promise<void> {
  const db = getDbInstance();

  // Hash input (SHA-256) - never store raw prompts
  const inputHash = createHash("sha256")
    .update(JSON.stringify(input))
    .digest("hex");

  // Truncate output to 200 chars
  const outputTruncated = JSON.stringify(output).slice(0, 200);

  const stmt = db.prepare(`
    INSERT INTO mcp_tool_audit (tool_name, input_hash, output_truncated, duration_ms, api_key_id, created_at)
    VALUES (?, ?, ?, ?, ?, ?)
  `);

  stmt.run(toolName, inputHash, outputTruncated, durationMs, apiKeyId, new Date().toISOString());
}
```

**Key Points**:
- SHA-256 hash input (never store raw prompts)
- Truncate output to 200 chars
- Store tool name, duration, API key ID
- SQLite persistence

#### 5.5 MCP Tools (29 total)

**Phase 1: Essential Tools (8)**
1. `omniroute_get_health` - Gateway health, uptime, memory, circuit breakers
2. `omniroute_list_combos` - List all combos with strategies and metrics
3. `omniroute_get_combo_metrics` - Performance metrics for specific combo
4. `omniroute_switch_combo` - Activate/deactivate combo for routing
5. `omniroute_check_quota` - Remaining API quota per provider
6. `omniroute_route_request` - Send chat completion through routing
7. `omniroute_cost_report` - Cost report by period
8. `omniroute_list_models_catalog` - List all available models

**Phase 2: Advanced Tools (21)**
9. `omniroute_simulate_route` - Dry-run routing simulation
10. `omniroute_set_budget_guard` - Set session budget guard
11. `omniroute_set_resilience_profile` - Apply resilience profile
12. `omniroute_test_combo` - Test each provider in combo
13. `omniroute_get_provider_metrics` - Per-provider metrics
14. `omniroute_best_combo_for_task` - AI-powered combo recommendation
15. `omniroute_explain_route` - Explain routing decision
16. `omniroute_get_session_snapshot` - Full session snapshot
17-29. Additional tools (cache, memory, skills, etc.)

---

## 6. Services Pattern

### Location: `open-sse/services/`

### Key Services (36+)

#### 6.1 Rate Limit Manager
```typescript
// rateLimitManager.ts - Token bucket rate limiting
class TokenBucket {
  capacity: number;
  tokens: number;
  lastRefill: number;

  constructor(capacity: number, refillRate: number) {
    this.capacity = capacity;
    this.tokens = capacity;
    this.lastRefill = Date.now();
  }

  consume(tokens: number): boolean {
    this.refill();
    if (this.tokens >= tokens) {
      this.tokens -= tokens;
      return true;
    }
    return false;
  }

  refill(): void {
    const now = Date.now();
    const elapsed = now - this.lastRefill;
    const tokensToAdd = (elapsed / 1000) * this.refillRate;
    this.tokens = Math.min(this.capacity, this.tokens + tokensToAdd);
    this.lastRefill = now;
  }
}

export function checkRateLimit(
  apiKey: string,
  provider: string
): { allowed: boolean; retryAfter?: number } {
  const bucket = getOrCreateBucket(apiKey, provider);
  return bucket.consume(1) ? { allowed: true } : { allowed: false, retryAfter: 60 };
}
```

**Key Points**:
- Token bucket algorithm
- Per API key + provider bucket
- Reject requests exceeding limits
- Return retry-after header

#### 6.2 Token Refresh
```typescript
// tokenRefresh.ts - OAuth token refresh
export async function refreshTokenIfNeeded(
  credentials: ProviderCredentials,
  provider: string
): Promise<ProviderCredentials> {
  if (!needsRefresh(credentials)) {
    return credentials;
  }

  const refreshed = await refreshOAuthToken(provider, credentials.refreshToken);

  return {
    ...credentials,
    accessToken: refreshed.accessToken,
    refreshToken: refreshed.refreshToken,
    expiresAt: refreshed.expiresAt,
  };
}

function needsRefresh(credentials: ProviderCredentials): boolean {
  if (!credentials.expiresAt) return false;
  const expiresAtMs = new Date(credentials.expiresAt).getTime();
  const now = Date.now();
  return expiresAtMs - now < 5 * 60 * 1000; // Refresh if < 5 min remaining
}
```

**Key Points**:
- Detect token expiration
- Refresh via OAuth endpoint
- Update credentials in-place
- Persist refreshed tokens

#### 6.3 Account Fallback
```typescript
// accountFallback.ts - Account-level fallback
export async function selectAccount(
  provider: string,
  strategy: string
): Promise<string> {
  const connections = await getActiveConnections(provider);

  // Apply strategy
  switch (strategy) {
    case "round-robin":
      return selectRoundRobin(connections);
    case "least-used":
      return selectLeastUsed(connections);
    case "priority":
      return selectByPriority(connections);
    default:
      return connections[0].id;
  }
}
```

**Key Points**:
- Select account based on strategy
- Track account usage
- Fallback to alternate account
- Load balancing across accounts

---

## 7. API Route Pattern

### Location: `src/app/api/v1/`

### Pattern Components

#### 7.1 Route Handler Pattern
```typescript
// src/app/api/v1/chat/completions/route.ts
import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";
import { handleChat } from "@/open-sse/handlers/chatCore";

const requestSchema = z.object({
  model: z.string(),
  messages: z.array(z.object({
    role: z.enum(["user", "assistant", "system"]),
    content: z.string(),
  })),
  stream: z.boolean().optional().default(false),
  temperature: z.number().optional(),
  max_tokens: z.number().optional(),
});

export async function POST(request: NextRequest) {
  // 1. CORS preflight
  const corsHeaders = {
    "Access-Control-Allow-Origin": "*",
    "Access-Control-Allow-Methods": "POST, OPTIONS",
    "Access-Control-Allow-Headers": "Content-Type, Authorization",
  };

  if (request.method === "OPTIONS") {
    return new NextResponse(null, { headers: corsHeaders });
  }

  // 2. Body validation (Zod)
  const body = await request.json();
  const validated = requestSchema.parse(body);

  // 3. Optional auth
  const apiKey = request.headers.get("authorization")?.replace("Bearer ", "");
  if (REQUIRE_API_KEY && !apiKey) {
    return NextResponse.json({ error: "API key required" }, { status: 401 });
  }

  // 4. API key policy enforcement
  const policy = await enforceApiKeyPolicy(apiKey, validated.model);
  if (!policy.allowed) {
    return NextResponse.json({ error: "Model not allowed" }, { status: 403 });
  }

  // 5. Prompt injection guard (clone request)
  const clonedRequest = cloneRequest(validated);
  const isInjection = detectPromptInjection(clonedRequest.messages);
  if (isInjection) {
    return NextResponse.json({ error: "Prompt injection detected" }, { status: 400 });
  }

  // 6. Handler delegation (open-sse)
  try {
    const response = await handleChat(validated, apiKey);
    return NextResponse.json(response, { headers: corsHeaders });
  } catch (error) {
    return NextResponse.json({ error: error.message }, { status: 500 });
  }
}
```

**Key Points**:
- CORS preflight handling
- Zod body validation
- Optional auth (controlled by env)
- API key policy enforcement
- Prompt injection guard (clones request)
- Handler delegation to open-sse
- Proper error handling

---

## 8. Key Takeaways

### Best Practices

1. **Database Layer**:
   - Domain modules own specific tables
   - No raw SQL in routes
   - Versioned migrations (idempotent)
   - Encryption at rest for sensitive fields
   - WAL journaling for concurrent reads

2. **Executor Pattern**:
   - Strategy pattern for provider-specific logic
   - Base class with common logic
   - Fallback mechanism (multiple base URLs)
   - Retry with exponential backoff
   - Auto credential refresh

3. **Combo Routing**:
   - 13 routing strategies
   - Combo-first design (all routing through combo)
   - Circuit breaker integration
   - Fallback chains
   - Cost-aware, latency-aware, health-aware

4. **Translator Pattern**:
   - OpenAI as canonical format
   - Normalize → Translate → Denormalize
   - Preserve cache control for Claude Code
   - Tool call normalization
   - Role normalization

5. **MCP Server**:
   - Zod schema validation
   - Scope-based access control
   - Audit logging (SHA-256 hash)
   - 29 tools across 2 phases
   - 3 transports (stdio, SSE, HTTP)

6. **API Routes**:
   - CORS preflight
   - Zod validation
   - Optional auth
   - Policy enforcement
   - Prompt injection guard
   - Handler delegation

### Architecture Patterns

1. **Layered Architecture**: API → Handlers → Executors → Services → DB
2. **Strategy Pattern**: Executors, routing strategies
3. **Factory Pattern**: Executor creation
4. **Singleton Pattern**: DB instance
5. **Repository Pattern**: Domain modules
6. **Observer Pattern**: Circuit breaker, rate limit events
7. **Command Pattern**: MCP tools
8. **Chain of Responsibility**: Combo fallback chain

### Performance Patterns

1. **Caching Everywhere**: Models, providers, quotas pre-cached
2. **Read Cache**: In-memory cache for read-heavy ops
3. **WAL Journaling**: Concurrent reads during writes
4. **Token Bucket**: Efficient rate limiting
5. **Connection Pooling**: Singleton DB instance
6. **Async Non-blocking**: No blocking I/O in hot path

### Security Patterns

1. **Encryption at Rest**: AES-256-GCM for credentials
2. **Scope Enforcement**: Fine-grained MCP scopes
3. **Input Validation**: Zod schemas everywhere
4. **Prompt Injection Guard**: Clone and detect
5. **Audit Logging**: SHA-256 hash, no raw prompts
6. **API Key Policies**: Model-level permissions

---

## 9. Applied to Ti Router

### Current Ti Router Architecture

```
apps/router/
├── cmd/routerd/
│   ├── server.go - HTTP server
│   └── handlers/
│       ├── admin/ - Admin endpoints
│       └── chat/ - Chat completion
├── layers/
│   ├── authentication/ - Auth layer
│   ├── provider/ - Provider layer
│   ├── routing/ - Routing layer
│   └── resilience/ - Resilience layer
└── ui/ - React UI
```

### Patterns to Apply

#### 9.1 Database Layer
**Current**: Go với direct SQL queries
**OmniRoute Pattern**: Domain modules với migrations
**Apply**:
- Create domain modules in Go (`pkg/db/`)
- Add migration runner
- Use prepared statements
- Encrypt sensitive fields

#### 9.2 Executor Pattern
**Current**: Provider-specific logic mixed in handlers
**OmniRoute Pattern**: Strategy pattern với BaseExecutor
**Apply**:
- Create `BaseExecutor` interface
- Implement provider-specific executors
- Add executor factory
- Separate common vs specific logic

#### 9.3 Combo Routing
**Current**: Simple model-to-provider mapping
**OmniRoute Pattern**: 13 routing strategies
**Apply**:
- Implement combo config structure
- Add routing strategies (priority, weighted, etc.)
- Create combo resolution logic
- Add fallback chains

#### 9.4 Translator Pattern
**Current**: Limited format support
**OmniRoute Pattern**: OpenAI as canonical format
**Apply**:
- Implement format registry
- Add request/response translators
- Normalize to canonical format
- Handle tool calls, roles, cache control

#### 9.5 MCP Server
**Current**: Basic MCP implementation
**OmniRoute Pattern**: 29 tools, 10 scopes, audit logging
**Apply**:
- Add more tools (health, quota, cost report)
- Implement scope enforcement
- Add audit logging
- Support multiple transports

---

## 10. Future Enhancements

### Short Term
1. Add domain modules to Ti Router DB layer
2. Implement BaseExecutor pattern
3. Add basic combo routing (priority strategy)
4. Implement request/response translators
5. Add MCP tools for health, quota, usage

### Medium Term
1. Add more routing strategies (weighted, round-robin)
2. Implement circuit breaker integration
3. Add token bucket rate limiting
4. Implement OAuth token refresh
5. Add account-level fallback

### Long Term
1. Implement all 13 routing strategies
2. Add auto-routing (AI-powered)
3. Implement context-aware routing
4. Add full MCP tool set (29 tools)
5. Implement Electron desktop app

---

## 11. References

- **OmniRoute**: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\07_Repositories\router\OmniRoute\`
- **OmniRoute GitHub**: https://github.com/diegosouzapw/OmniRoute
- **Ti Router**: `Z:\10_WORKPLACE\Ti\apps\router\`
- **MCP Protocol**: https://modelcontextprotocol.io/
- **better-sqlite3**: https://github.com/WiseLibs/better-sqlite3
- **Zod**: https://zod.dev/
