# AI Agents Configuration & Status

This index tracks all 6 core AI agents and their current implementation status.

## Core Pipeline (Active)

| Agent ID | Tier | Role | Spec | Prompt | Impl |
|----------|------|------|------|--------|------|
| [architect](./architect-spec.md) | T1 | Intent & Risk Analysis | ✅ | ✅ | ✅ |
| [planner](./planner-spec.md) | T2B | Execution Planning | ✅ | ✅ | ✅ |
| [coder](./coder-spec.md) | T2A | Code Generation | ✅ | ✅ | ✅ |
| [validator](./validator-spec.md) | T2A | Semantic Validation | ⏳ | ⏳ | ⏳ |
| [guard](./guard-spec.md) | T2B | Security Policy | ⏳ | ⏳ | ⏳ |
| [summarizer](./summarizer-spec.md) | T2A | UX Summaries | ⏳ | ⏳ | ⏳ |

## Reference

Original specs derived from `docs/AGENT_SPECS.md`.

## Integration Status

- **Orchestrator**: Implemented (`packages/spec/src/agents/orchestrator.ts`)
- **Runtime**: Integration pending
- **VS Code**: Integration pending
