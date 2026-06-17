# CLI Knowledge Base Index

> **Last Updated**: 2026-05-05 (updated 2026-05-05)
> **Purpose**: Navigation guide for Ti CLI documentation

---

## Overview

This directory contains documentation specifically for the Ti CLI project. Non-CLI related documentation has been moved to appropriate folders in the parent directory.

---

## Folder Structure

### 📁 core - Core CLI Architecture

Core architecture and fundamental concepts of the Ti CLI.

- **[ARCHITECTURE.md](core/ARCHITECTURE.md)** - Overall CLI architecture
- **[PLUGIN_ARCHITECTURE.md](core/PLUGIN_ARCHITECTURE.md)** - Plugin system architecture
- **[QUICK_START.md](core/QUICK_START.md)** - Quick start guide
- **[CLI_GUIDE.md](core/CLI_GUIDE.md)** - Comprehensive CLI guide
- **[CODEBASE_MIGRATION_ASSESSMENT.md](core/CODEBASE_MIGRATION_ASSESSMENT.md)** - Codebase migration assessment

### 🔧 features - CLI Features

Documentation about specific CLI features and capabilities.

- **[CLI_TOOL_SYSTEM.md](features/CLI_TOOL_SYSTEM.md)** - Tool system documentation
- **[CLI_COMMAND_VALIDATION.md](features/CLI_COMMAND_VALIDATION.md)** - Command validation
- **[CLI_ERROR_HANDLING.md](features/CLI_ERROR_HANDLING.md)** - Error handling
- **[CLI_FLAG_MANAGEMENT.md](features/CLI_FLAG_MANAGEMENT.md)** - Flag management
- **[CLI_EXAMPLES.md](features/CLI_EXAMPLES.md)** - Usage examples

### 🔌 integration - CLI Integration

Documentation about integrating the CLI with other systems.

- **[CLI_INTEGRATION_ROADMAP.md](integration/CLI_INTEGRATION_ROADMAP.md)** - CLI integration roadmap
- **[CLI_MCP_OAUTH.md](integration/CLI_MCP_OAUTH.md)** - CLI MCP OAuth integration
- **[TI_CLI_BEST_SOURCE_INTEGRATION.md](integration/TI_CLI_BEST_SOURCE_INTEGRATION.md)** - Integration with best_source (dynamic read)
- **[TI_CLI_CLAUDE_FORMAT.md](integration/TI_CLI_CLAUDE_FORMAT.md)** - Integration with Claude Code format

### 🚀 deployment - CLI Deployment

Deployment and maintenance documentation for the CLI.

- **[DEPLOYMENT_MAINTENANCE_STRATEGY.md](deployment/DEPLOYMENT_MAINTENANCE_STRATEGY.md)** - Deployment and maintenance strategy

### 📚 sources - CLI Source Analysis

Analysis of other CLI implementations for learning and reference.

- **[chatgpt-cli/](sources/chatgpt-cli/)** - ChatGPT CLI analysis
  - **[CHATGPT_CLI_PLUGIN_SEPARATION_ANALYSIS.md](sources/chatgpt-cli/CHATGPT_CLI_PLUGIN_SEPARATION_ANALYSIS.md)** - Plugin separation analysis
  - **[README.md](sources/chatgpt-cli/README.md)** - ChatGPT CLI source information

---

## Related Documentation

Documentation for related systems has been moved to appropriate folders:

### Agent Framework
- **Location**: `../agents/`
- **Contents**: Agent architecture, TI Crew, sub-agent guides

### Router
- **Location**: `../Router/`
- **Contents**: Router middleware architecture, rules enforcement

### Notion Integration
- **Location**: `../notion/`
- **Contents**: Notion sync strategy, architecture, integration points

### Research
- **Location**: `../research/`
- **Contents**: CLI upgrade research, patterns, analysis of other CLIs

### Devin Porting
- **Location**: `../devin/`
- **Contents**: Devin API architecture, Go port plans

### Lessons Learned
- **Location**: `../lessons/`
- **Contents**: Lessons learned from other CLI implementations

---

## Quick Navigation

### For New Contributors
1. Start with **[core/QUICK_START.md](core/QUICK_START.md)**
2. Read **[core/ARCHITECTURE.md](core/ARCHITECTURE.md)**
3. Explore **[core/CLI_GUIDE.md](core/CLI_GUIDE.md)**

### For Plugin Development
1. Read **[core/PLUGIN_ARCHITECTURE.md](core/PLUGIN_ARCHITECTURE.md)**
2. Check **[features/](features/)** for feature examples
3. Review **[integration/](integration/)** for integration patterns

### For Deployment
1. Review **[deployment/DEPLOYMENT_MAINTENANCE_STRATEGY.md](deployment/DEPLOYMENT_MAINTENANCE_STRATEGY.md)**

---

## Maintenance

When adding new CLI documentation:
1. Choose the appropriate folder based on topic (core, features, integration, deployment, sources)
2. Add an entry to this index
3. Use clear, descriptive filenames
4. Update this index's "Last Updated" date

**Important**: Only add CLI-specific documentation to this folder. Move non-CLI content to appropriate folders in the parent directory.

---

## Related Resources

- **Main Ti Repository**: `Z:\10_WORKPLACE\Ti\`
- **CLI Source**: `apps/cli/` in the Ti monorepo
- **Learning Lab**: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\`
- **Agent Documentation**: `../agents/`
- **Router Documentation**: `../Router/`
- **Research**: `../research/`
