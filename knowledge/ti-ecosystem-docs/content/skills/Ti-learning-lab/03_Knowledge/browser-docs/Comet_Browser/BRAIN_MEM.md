# Perplexity AI & Comet Browser Knowledge Graph

## Architecture
- **Comet Browser**: Fork of Chromium with embedded AI agent capabilities.
- **Agentic Extension**: Hidden extension (gents.crx) enabling MCP (Model Context Protocol).
- **Key API**: chrome.perplexity.mcp.addStdioServer (allows local process execution).
- **Communication**: Uses externally_connectable to listen to perplexity.ai (and now localhost).

## Capabilities (Unlocked)
1. **Local Execution**: Can spawn cmd.exe, powershell, or any local tool.
2. **Screen Vision**: Can capture screenshots and send to LLM.
3. **DOM Extraction**: Optimized for extracting clean text/PDF content.

## Integration Strategy (Dual-Stack)
- **Primary**: Perplexity Cloud controls the agent via perplexity.ai.
- **Secondary**: Spectre/Jarvis controls the agent via localhost bridge.
- **Patch**: Modified manifest.json to whitelist localhost.

## Artifacts
- **Extension Path**: Z:\knowledge_base\04_Research\00_Active_Research\Comet_Browser\agents_extension
- **Manifest**: Patched to include http://localhost/*.
- **Background Script**: ackground.js (Analyzed).

## Risks
- Unsandboxed execution of local commands.
- Hidden from standard extension management.
