# Auth Directory Setup

The `auth/` directory is used by CLIProxyAPI for storing OAuth tokens and authentication files.

## OAuth Providers

CLIProxyAPI supports OAuth for:
- Google Gemini
- OpenAI Codex
- Claude Code
- Qwen Code
- iFlow
- Antigravity

## Usage

OAuth tokens are automatically stored in the `auth/` directory when you run:
```bash
cliproxyapi.exe -login          # Gemini login
cliproxyapi.exe -codex-login    # Codex login
cliproxyapi.exe -claude-login   # Claude login
```

## Manual Token Import

For manual token import, use the scripts in `Z:\02_CORE\_cli\config\`:
- `import-oauth-tokens.ps1`
- `load-tokens-from-auth.ps1`

## Security

- The `auth/` directory should not be committed to git
- Tokens are stored securely by CLIProxyAPI
- See `Tiserverrouter.yaml` for OAuth configuration
