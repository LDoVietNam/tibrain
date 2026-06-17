# Planner Agent Specification

**Agent ID**: `planner`  
**Tier**: T2B (Supreme Planner)  
**Status**: Active  
**Version**: 1.0.0

## Mission
Create detailed, step-by-step execution plans based on architectural analysis. Ensures changes are atomic, ordered correctly, and dependencies are managed.

## Capabilities

### Must-Have
- Step-by-step plan generation
- File dependency resolution
- Complexity estimation
- Warning generation for potential issues

### Nice-to-Have
- Parallel execution paths
- Rollback strategy per step

## Input Contract
```typescript
interface PlannerInput {
  architectOutput: ArchitectOutput;
  context?: {
    codebase?: string[];
    dependencies?: string[];
  };
}
```

## Output Contract
```typescript
interface PlannerOutput {
  plan: {
    steps: Array<{
      id: number;
      action: string;
      file: string;
      description: string;
      dependencies: number[];
    }>;
    estimated_files: number;
    estimated_complexity: 'low' | 'medium' | 'high';
  };
  files_to_modify: string[];
  files_to_read: string[];
  warnings: string[];
}
```

## AI Configuration
- **Models**: GPT-4, Claude 3 Opus
- **Temperature**: 0.2
- **System Prompt**: `.ai/prompts/agents/planner.md`

## Integration
- **Upstream**: Architect Agent
- **Downstream**: Coder Agent
