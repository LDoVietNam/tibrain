# TiBrain Data Table Catalog

This is the shortest path to the important tables in `TiBrain`.

## Search First

- Document search: `rag_documents_fts`
- Memory search: `memory_fts`
- Skill lookup: `skills.name`
- API key lookup: `api_keys.key`
- Recent activity: `call_logs.timestamp`, `usage_history.timestamp`

## RAG And Knowledge

| Table | Purpose | Best search keys |
|---|---|---|
| `rag_documents` | Indexed knowledge documents | `title`, `content`, `path`, `category`, `tags`, `status`, `updated_at` |
| `rag_knowledge_bases` | Knowledge base registry | `name`, `path`, `status`, `updated_at` |
| `rag_query_history` | Query history | `query`, `user_id`, `session_id`, `timestamp` |
| `rag_runtime_traces` | Retrieval runtime trace log for verify/score/distill stages | `query_text`, `mode`, `verified`, `quality_score`, `created_at` |
| `brain_pattern_candidates` | Distilled promotion candidates and promoted runtime patterns | `query_id`, `status`, `score`, `promoted_at`, `updated_at` |
| `rag_vector_index` | Chunk/vector mapping | `document_id`, `chunk_id` |
| `rag_analytics` | RAG metrics | `metric_type`, `metric_name`, `timestamp` |
| `rag_feedback` | User feedback on queries | `query_id`, `rating`, `timestamp` |

## Memory And Context

| Table | Purpose | Best search keys |
|---|---|---|
| `memories` | Long-lived memory store | `api_key_id`, `session_id`, `type`, `key`, `expires_at` |
| `memory_fts` | Full-text search for memory content | `content`, `key` |
| `cli_context` | CLI-scoped context storage | `cli_id`, `context_type`, `key`, `expires_at` |
| `cli_query_cache` | Cached query responses | `query_hash`, `source_type`, `expires_at` |
| `cli_help_index` | Help article index | `command`, `topic`, `category`, `keywords` |

## Routing And Provider Operations

| Table | Purpose | Best search keys |
|---|---|---|
| `provider_connections` | Provider auth and health state | `provider`, `is_active`, `priority`, `name` |
| `provider_nodes` | Provider node registry | `type`, `name`, `api_type`, `base_url` |
| `combos` | Combo definitions | `name`, `created_at` |
| `call_logs` | Request-level log summary | `timestamp`, `status`, `requested_model`, `request_type`, `combo_name` |
| `proxy_logs` | Proxy execution log | `timestamp`, `status`, `provider` |
| `usage_history` | Token and latency history | `timestamp`, `provider`, `model` |
| `semantic_cache` | Prompt/response cache | `signature`, `model`, `expires_at` |
| `reasoning_cache` | Replayed reasoning cache | `tool_call_id`, `provider`, `model`, `expires_at` |

## Coordination And Registry

| Table | Purpose | Best search keys |
|---|---|---|
| `api_keys` | API key registry | `key`, `name`, `machine_id` |
| `skills` | Skill definitions | `name`, `api_key_id`, `enabled` |
| `skill_executions` | Skill execution log | `skill_id`, `api_key_id`, `status`, `created_at` |
| `files` | Uploaded file metadata | `filename`, `purpose`, `api_key_id`, `status` |
| `batches` | Batch job tracking | `status`, `api_key_id`, `model`, `created_at` |
| `mcp_tool_audit` | MCP tool audit log | `tool_name`, `api_key_id`, `created_at` |
| `a2a_tasks` | A2A task lifecycle | `state`, `skill_id`, `api_key_id`, `created_at` |
| `routing_decision_log` | Retrieval-router learning log | `query`, `route`, `confidence`, `rule_matched`, `timestamp` |
| `routing_decisions` | Provider-routing explainability | `request_id`, `combo_id`, `provider_selected`, `created_at` |
| `combo_adaptation_state` | Learned combo scores | `combo_id`, `provider_id`, `updated_at` |

## Operational Notes

- For document browsing, use `GET /api/v1/knowledge/documents?q=...`.
- For content search, `q` now uses the document FTS index instead of scanning only by category.
- The learning plane persists its evidence in `rag_runtime_traces` and `brain_pattern_candidates`.
- The execution plane should read only `brain_pattern_candidates` rows with `status='promoted'`.
- Prefer `updated_at DESC` for recency views and `status=active` for live data.
- If you need to add a new searchable table, start by defining the access pattern first, then add a dedicated index or FTS table.
