# Database Schema - master_database.db

**Location:** `Z:\knowledge_base\06_database\master_database.db`

---

## Core Tables

### provider_connections

Stores all provider OAuth/API key connections.

```sql
CREATE TABLE provider_connections (
  id TEXT PRIMARY KEY,
  provider TEXT NOT NULL,              -- 'iflow', 'qwen', 'codex', etc.
  auth_type TEXT,                       -- 'oauth', 'apikey'
  name TEXT,                            -- User-defined name
  email TEXT,                           -- OAuth account email
  priority INTEGER DEFAULT 0,           -- Routing priority
  is_active INTEGER DEFAULT 1,          -- Connection active flag

  -- OAuth tokens
  access_token TEXT,
  refresh_token TEXT,
  expires_at TEXT,
  token_expires_at TEXT,
  scope TEXT,
  token_type TEXT,
  id_token TEXT,
  expires_in INTEGER,

  -- API key
  api_key TEXT,

  -- Provider metadata
  project_id TEXT,
  provider_specific_data TEXT,

  -- Health & testing
  test_status TEXT,                     -- 'success', 'failed', 'pending'
  error_code TEXT,
  last_error TEXT,
  last_error_at TEXT,
  last_error_type TEXT,
  last_error_source TEXT,
  last_tested TEXT,

  -- Rate limiting & backoff
  backoff_level INTEGER DEFAULT 0,
  rate_limited_until TEXT,
  rate_limit_protection INTEGER DEFAULT 0,
  health_check_interval INTEGER,
  last_health_check_at TEXT,

  -- Display
  display_name TEXT,
  global_priority INTEGER,
  default_model TEXT,

  -- Usage tracking
  consecutive_use_count INTEGER DEFAULT 0,

  -- Timestamps
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
```

**Current Data (2026-03-30):**
| Provider | Count | Status |
|----------|-------|--------|
| iflow | 8 | expired (refresh_failed) |
| qwen | 13 | active |
| codex | 5 | active |
| antigravity | 7 | active |
| gemini | 3 | active |
| kimi | 3 | active |
| claude-oauth | 1 | active |

---

### provider_nodes

Defines OpenAI-compatible provider endpoints.

```sql
CREATE TABLE provider_nodes (
  id TEXT PRIMARY KEY,
  type TEXT NOT NULL,                   -- 'openai-compatible', 'anthropic-compatible'
  name TEXT NOT NULL,                   -- Display name
  prefix TEXT,                          -- Model prefix (e.g., 'openai-compatible-')
  api_type TEXT,                        -- API type identifier
  base_url TEXT,                        -- Base URL for API
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
```

**Current Data:** EMPTY - Needs configuration for iflow, qwen, kimi, antigravity

**Example Insert:**
```sql
INSERT INTO provider_nodes (id, type, name, prefix, api_type, base_url) VALUES
('node_iflow', 'openai-compatible', 'iFlow AI', 'openai-compatible-', 'openai', 'https://api.iflow.io/v1'),
('node_qwen', 'openai-compatible', 'Qwen Code', 'openai-compatible-', 'openai', 'https://api.qwen.ai/v1'),
('node_kimi', 'openai-compatible', 'Kimi', 'openai-compatible-', 'openai', 'https://api.kimi.moonshot.cn/v1'),
('node_antigravity', 'openai-compatible', 'Antigravity', 'openai-compatible-', 'openai', 'https://api.antigravity.dev/v1');
```

---

### key_value

Key-value store for settings, pricing, proxy config.

```sql
CREATE TABLE key_value (
  namespace TEXT NOT NULL,              -- 'settings', 'pricing', 'proxyConfig'
  key TEXT NOT NULL,
  value TEXT NOT NULL,                  -- JSON string
  PRIMARY KEY (namespace, key)
);
```

**Current Data:**
| namespace | key | value |
|-----------|-----|-------|
| settings | password | "$2b$10$JQ5Hjh8GZsG7DpymA1tNqOFyaZrqQFzDRbU3DcUXbq2p.dhnGZjWS" |
| settings | requireLogin | true |
| settings | setupComplete | true |

---

### combos

Model combination definitions for routing.

```sql
CREATE TABLE combos (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,            -- Combo name
  data TEXT NOT NULL,                   -- JSON: { models: [...], strategy: '...' }
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
```

**data JSON Schema:**
```json
{
  "models": [
    { "provider": "iflow", "model": "deepseek-v3.2" },
    { "provider": "qwen", "model": "qwen-max" }
  ],
  "strategy": "round-robin",            -- 'fill-first', 'round-robin', 'p2c', 'random', 'least-used', 'cost-optimized'
  "settings": {
    "stickyLimit": 3,
    "timeout": 30000
  }
}
```

---

### api_keys

API keys for external clients.

```sql
CREATE TABLE api_keys (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  key TEXT NOT NULL UNIQUE,             -- Hashed key
  machine_id TEXT,                      -- Bound machine ID
  allowed_models TEXT,                  -- JSON: ["iflow/*", "qwen/*"] or ["*"]
  created_at TEXT NOT NULL
);
```

---

## Logging Tables

### call_logs

Request/response logs.

```sql
CREATE TABLE call_logs (
  id TEXT PRIMARY KEY,
  timestamp TEXT NOT NULL,
  method TEXT,
  path TEXT,
  status INTEGER,
  model TEXT,
  provider TEXT,
  account TEXT,
  connection_id TEXT,
  duration INTEGER DEFAULT 0,           -- ms
  tokens_in INTEGER DEFAULT 0,
  tokens_out INTEGER DEFAULT 0,
  source_format TEXT,
  target_format TEXT,
  api_key_id TEXT,
  api_key_name TEXT,
  combo_name TEXT,
  request_body TEXT,                    -- JSON
  response_body TEXT,                   -- JSON
  error TEXT
);
```

---

### proxy_logs

Proxy routing logs.

```sql
CREATE TABLE proxy_logs (
  id TEXT PRIMARY KEY,
  timestamp TEXT NOT NULL,
  status TEXT,
  proxy_type TEXT,
  proxy_host TEXT,
  proxy_port INTEGER,
  level TEXT,                           -- 'global', 'provider', 'combo', 'key'
  level_id TEXT,
  provider TEXT,
  target_url TEXT,
  public_ip TEXT,
  latency_ms INTEGER DEFAULT 0,
  error TEXT,
  connection_id TEXT,
  combo_id TEXT,
  account TEXT,
  tls_fingerprint INTEGER DEFAULT 0
);
```

---

### usage_history

Token usage tracking.

```sql
CREATE TABLE usage_history (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  provider TEXT,
  model TEXT,
  connection_id TEXT,
  api_key_id TEXT,
  api_key_name TEXT,
  tokens_input INTEGER DEFAULT 0,
  tokens_output INTEGER DEFAULT 0,
  tokens_cache_read INTEGER DEFAULT 0,
  tokens_cache_creation INTEGER DEFAULT 0,
  tokens_reasoning INTEGER DEFAULT 0,
  status TEXT,
  timestamp TEXT NOT NULL
);
```

---

### audit_log

Audit trail for compliance.

```sql
CREATE TABLE audit_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  timestamp TEXT NOT NULL DEFAULT (datetime('now')),
  action TEXT NOT NULL,
  actor TEXT NOT NULL DEFAULT 'system',
  target TEXT,
  details TEXT,                         -- JSON
  ip_address TEXT
);
```

---

## Domain Management Tables

### domain_fallback_chains

Fallback chain configuration per model.

```sql
CREATE TABLE domain_fallback_chains (
  model TEXT PRIMARY KEY,
  chain TEXT NOT NULL                   -- JSON array: ["provider1", "provider2"]
);
```

**Example:**
```sql
INSERT INTO domain_fallback_chains (model, chain) VALUES
('gpt-4', '["openai","anthropic","deepseek","iflow"]'),
('claude-sonnet', '["anthropic","iflow","qwen"]');
```

---

### domain_budgets

API key budget tracking.

```sql
CREATE TABLE domain_budgets (
  api_key_id TEXT PRIMARY KEY,
  daily_limit_usd REAL NOT NULL,
  monthly_limit_usd REAL DEFAULT 0,
  warning_threshold REAL DEFAULT 0.8    -- 80% = warning
);
```

---

### domain_cost_history

Cost tracking over time.

```sql
CREATE TABLE domain_cost_history (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  api_key_id TEXT NOT NULL,
  cost REAL NOT NULL,
  timestamp INTEGER NOT NULL            -- Unix timestamp
);
```

---

### domain_lockout_state

Rate limit lockout management.

```sql
CREATE TABLE domain_lockout_state (
  identifier TEXT PRIMARY KEY,          -- IP or API key ID
  attempts TEXT NOT NULL,               -- JSON array of timestamps
  locked_until INTEGER                  -- Unix timestamp
);
```

---

### domain_circuit_breakers

Circuit breaker state for resilience.

```sql
CREATE TABLE domain_circuit_breakers (
  name TEXT PRIMARY KEY,                -- Provider or combo ID
  state TEXT NOT NULL DEFAULT 'CLOSED', -- 'CLOSED', 'OPEN', 'HALF-OPEN'
  failure_count INTEGER DEFAULT 0,
  last_failure_time INTEGER,            -- Unix timestamp
  options TEXT                          -- JSON config
);
```

---

## Cache Tables

### semantic_cache

Response cache for repeated prompts.

```sql
CREATE TABLE semantic_cache (
  id TEXT PRIMARY KEY,
  signature TEXT NOT NULL UNIQUE,       -- Semantic signature hash
  model TEXT NOT NULL,
  prompt_hash TEXT NOT NULL,
  response TEXT NOT NULL,               -- Cached response
  tokens_saved INTEGER DEFAULT 0,
  hit_count INTEGER DEFAULT 0,
  created_at TEXT NOT NULL,
  expires_at TEXT NOT NULL
);
```

---

## Meta Tables

### settings

Legacy settings table (use key_value instead).

```sql
CREATE TABLE settings (
  key TEXT PRIMARY KEY,
  value TEXT
);
```

---

### db_meta

Database metadata.

```sql
CREATE TABLE db_meta (
  key TEXT PRIMARY KEY,
  value TEXT
);
```

---

### _tiroute_migrations

Migration tracking.

```sql
CREATE TABLE _tiroute_migrations (
  version TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  applied_at TEXT NOT NULL DEFAULT (datetime('now'))
);
```

---

### sqlite_sequence

SQLite auto-increment tracking (system table).

```sql
CREATE TABLE sqlite_sequence(name,seq);
```

---

## Queries Reference

### Get Provider Connection Count
```sql
SELECT provider, COUNT(*) as count
FROM provider_connections
GROUP BY provider;
```

### Get Expired Connections
```sql
SELECT id, provider, email, token_expires_at
FROM provider_connections
WHERE token_expires_at < datetime('now')
  AND auth_type = 'oauth';
```

### Get Usage by Provider (Last 24h)
```sql
SELECT provider,
       SUM(tokens_input) as input_tokens,
       SUM(tokens_output) as output_tokens,
       COUNT(*) as requests
FROM usage_history
WHERE timestamp >= datetime('now', '-24 hours')
GROUP BY provider;
```

### Get Failed Requests
```sql
SELECT model, provider, error, COUNT(*) as count
FROM call_logs
WHERE status >= 400
GROUP BY model, provider, error
ORDER BY count DESC;
```

---

## Current State Summary

| Table | Records | Status |
|-------|---------|--------|
| provider_connections | 40 | ✓ Active |
| provider_nodes | 0 | ⚠ Empty |
| key_value (settings) | 3 | ✓ Active |
| combos | 0 | ⚠ Empty |
| domain_fallback_chains | 0 | ⚠ Empty |
| api_keys | - | - |
| call_logs | - | Logging |
| usage_history | - | Logging |
