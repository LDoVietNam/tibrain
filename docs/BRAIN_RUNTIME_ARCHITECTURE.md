# TiBrain Brain/Runtime Architecture

## Core Pipeline

TiBrain now treats knowledge evolution as a staged pipeline:

`observe -> verify -> score -> distill -> promote -> execute`

This is not just naming. Each stage has a distinct storage and API boundary.

## Two Planes

### Learning plane

Responsible for:

- ingesting docs, logs, sessions, and worker output
- retrieving broad evidence from the corpus
- verifying whether an answer is grounded
- scoring confidence, groundedness, and coverage
- distilling reusable pattern candidates
- proposing promotion into runtime

Learning-plane persistence:

- `rag_query_history`
- `rag_runtime_traces`
- `brain_pattern_candidates`

### Execution plane

Responsible for:

- consuming only promoted patterns
- turning promoted knowledge into runtime rules, workflows, contracts, and policy
- enforcing approval, safety, retry, and resume behavior
- rejecting raw or unverified knowledge

Execution-plane read boundary:

- `GET /api/v2/runtime/registry`
- `brain_pattern_candidates.status = 'promoted'`

## Design Rules

### 1. Workers cannot write logic directly

Worker or substrate output can be observed and indexed, but it must first pass:

- retrieval
- verification
- scoring
- distillation
- explicit promotion

Direct writes from worker output into runtime logic are intentionally blocked by architecture.

### 2. Retrieval is adaptive

The retrieval runtime uses:

- fast path for cheap lexical/vector retrieval
- slow path only when confidence is weak or evidence is thin
- runtime traces so every query leaves an inspectable decision trail

For local single-node operation, heavy paths are load-shed by default:

- vector retrieval is opt-in
- LLM answer refinement is opt-in
- embedding calls are skipped when there is no vector index to search
- API request handlers reuse a live RAG manager instance
- indexing batches knowledge-base stats updates instead of recomputing them per file
- SQLite uses a local-lean runtime profile with lighter sync and in-memory temp storage
- fresh databases are created directly at the latest `tool_registry` schema
- startup migrations are gated by `PRAGMA user_version` instead of re-checking every tool column on each boot

### 3. LLM output is verified before adoption

If an LLM answer is not grounded enough in retrieved evidence, TiBrain falls back to an extractive response before scoring and distillation.

This keeps the learning plane from promoting stylish but weak answers.

### 4. Promotion is the execution-plane gate

Pattern candidates exist in a holding area until an explicit promotion step marks them as runtime-safe enough to consume.

## API Surface

- `POST /api/v2/retrieve`
- `GET /api/v2/retrieve/traces`
- `GET /api/v2/brain/patterns`
- `POST /api/v2/brain/patterns/promote`
- `GET /api/v2/runtime/registry`

## Practical Reading Order

If you are extending this system, inspect in this order:

1. `adaptive_retrieval.go`
2. `api_server.go`
3. `rag_system.go`
4. `tag_rule_engine.go`
5. `retrieval_router.go`

That path follows the new boundary from learning-plane retrieval into execution-plane consumption.
