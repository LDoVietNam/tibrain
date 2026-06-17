# Changelog

## v0.7.0 - complete-translate-converter

- Completed `ti translate` as a practical converter layer rather than a shallow scanner.
- Added expanded Ti IR fields: workflows, hooks, plugins, providers, permissions, conflicts, and risks.
- Added source detection for Codex/OpenAI, Claude Code, OpenCode, Cursor, Windsurf, MCP, Aider, Roo, Devin-like, Kilo-like, and generic configs.
- Added export support to Claude Code, Codex, and OpenCode formats.
- Added `ti translate formats`, `ti translate conflicts`, `ti translate export`, and `ti translate optimize`.
- Improved frontmatter parsing, tool normalization, argument normalization, `$ARGUMENTS` conversion, MCP parsing, provider extraction, hook parsing, permission parsing, and conflict detection.
- Updated docs and translate examples.

## v0.6.0 - max-optimized-runtime

- Added `ti provider scan` and `ti provider import` for Kilo/OpenCode/generic provider configs.
- Added Ti Pack Manager for generated Ti packs.
- Added block search/stats, optimization diagnostics, shell completion, and max release scripts.

## v0.5.0 - translate-blocks-runtime

- Added Ti Blocks Runtime and block recording for sandbox/tool/translate runs.

## v0.4.0 - translate-engine

- Added first Ti Translate Engine, Ti IR / `ti.pack.yaml` output, and `compat.lock.json`.

## v0.8.0 - ultimate-optimizer-runtime

### Added

- Auth token manager with safe Windsurf token import helper.
- Token optimizer inspired by command-output proxy workflows.
- `ti-cli token estimate`, `token compress`, `token run`, and `token cache`.
- `ti-cli proxy run` compatibility alias.
- Token compression modes for git diff, test logs, install logs, JSON-like logs, and generic output.
- Cache and block metadata for token savings.
- Docs: `docs/TOKEN_OPTIMIZER.md`, `docs/AUTH_TOKENS.md`.
- Config sample: `configs/token_optimizer.example.json`.

### Security

- Direct token command-line argument import is intentionally not supported to avoid shell history leaks.
- Windsurf flow is manual/official-token-page based; Ti does not scrape or bypass provider authentication.
