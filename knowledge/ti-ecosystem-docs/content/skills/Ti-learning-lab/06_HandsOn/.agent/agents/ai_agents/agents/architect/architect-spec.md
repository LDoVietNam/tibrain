# Architect Agent Specification

**Agent ID**: `architect`  
**Tier**: T1  
**Status**: Active  
**Version**: 1.0.0

## Mission
Analyze user requests, assess risk, and define the architectural boundaries of changes. Acts as the first line of defense against scope creep and high-risk changes.

## Capabilities

### Must-Have
- Intent analysis (Refactor vs Feature vs Bugfix)
- Risk assessment (Low/Medium/High)
- Symbol identification (Files/Functions/Classes)
- Constraint definition (No logic change, UI only, etc.)

### Nice-to-Have
- Historical context awareness
- Dependency impact analysis

## Input Contract
```typescript
interface ArchitectInput {
  userRequest: string;
  context?: {
    currentFile?: string;
    selectedCode?: string;
    projectType?: string;
  };
}
```

## Output Contract
```typescript
interface ArchitectOutput {
  intent: {
    category: 'refactor' | 'feature' | 'bugfix' | 'style' | 'docs';
    description: string;
    scope: 'file' | 'module' | 'project';
  };
  risk: {
    level: 'low' | 'medium' | 'high';
    factors: string[];
    concerns: string[];
  };
  symbols: {
    files: string[];
    functions: string[];
    classes: string[];
  };
  constraints: {
    no_logic_change?: boolean;
    no_state_change?: boolean;
    ui_only?: boolean;
  };
}
```

## AI Configuration
- **Models**: GPT-4, Claude 3 Opus, Gemini Pro
- **Temperature**: 0.3
- **System Prompt**: `.ai/prompts/agents/architect.md`

## Integration
- **Upstream**: User Request
- **Downstream**: Planner Agent
