# Coder Agent Specification

**Agent ID**: `coder`  
**Tier**: T2A (Code Writer)  
**Status**: Active  
**Version**: 1.0.0

## Mission
Generate high-quality, minimal code changes in Unified Diff format. strictly adhering to the plan and constraints defined by upstream agents.

## Capabilities

### Must-Have
- Unified Diff generation
- Adherence to existing code style
- Minimal changes (surgical edits)
- JSON summary generation

### Nice-to-Have
- Syntax checking (basic)
- Import management

## Input Contract
```typescript
interface CoderInput {
  plannerOutput: PlannerOutput;
  fileContents: Map<string, string>;
  constraints?: {
    max_diff_bytes?: number;
    forbidden_patterns?: string[];
  };
}
```

## Output Contract
```typescript
interface CoderOutput {
  diff: string; // Unified diff format
  files_modified: string[];
  summary: {
    additions: number;
    deletions: number;
    files_changed: number;
    total_bytes: number;
  };
}
```

## AI Configuration
- **Models**: GPT-4, Claude 3 Opus, Gemini Pro
- **Temperature**: 0.1
- **System Prompt**: `.ai/prompts/agents/coder.md`

## Integration
- **Upstream**: Planner Agent
- **Downstream**: Validator Agent / Runtime Engine
