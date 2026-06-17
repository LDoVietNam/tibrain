# Ti CLI Max Optimization Track

This pack focuses on Ti as a translate-first, sandboxed AI workflow runtime.

## Optimization pillars

1. **Translate Engine** converts external AI-CLI formats into `ti.pack.yaml` instead of running foreign plugins directly.
2. **Provider Translate** imports Kilo/OpenCode/generic OpenAI-compatible provider config into Ti `compatible_providers` snippets.
3. **Sandbox Runtime** keeps shell/plugin/MCP execution behind policy, timeout, output limits, and audit logs.
4. **Blocks Runtime** stores command/output as replayable, searchable blocks.
5. **Pack Manager** lets generated packs be listed, inspected, enabled, or disabled.
6. **Release Optimization** favors Go 1.23 local toolchain, static small binaries, optional PGO, and smoke checks.

## New commands in this pack

```bash
ti-cli provider scan .
ti-cli provider import . --dry-run
ti-cli provider import ./opencode.json --out .ti/providers.generated.json

ti-cli pack list
ti-cli pack show .ti/packs/imported
ti-cli pack disable .ti/packs/imported
ti-cli pack enable .ti/packs/imported

ti-cli block search "error"
ti-cli block stats

ti-cli optimize doctor
ti-cli optimize build-plan

ti-cli completion bash
ti-cli completion zsh
ti-cli completion fish
ti-cli completion powershell
```

## Recommended max release build

```bash
GOTOOLCHAIN=local GOWORK=off go mod tidy
./scripts/max-release.sh VERSION=v0.6.0 OUT=bin/ti-cli
./bin/ti-cli version
./bin/ti-cli optimize doctor
./bin/ti-cli sandbox doctor
./bin/ti-cli translate scan .
```

## Provider Translate

The provider importer is heuristic and intentionally conservative. It looks for OpenAI-compatible fields such as:

- `base_url`, `baseURL`, `api_base`, `endpoint`, `url`
- `api_key_env`, `apiKeyEnv`, `env`
- `models`, `default_model`, `defaultModel`, `model`
- `headers`, `custom_headers`

It emits a Ti config snippet:

```json
{
  "compatible_providers": [
    {
      "name": "openrouter",
      "type": "openai_compatible",
      "base_url": "https://openrouter.ai/api/v1",
      "chat_path": "/chat/completions",
      "api_key_env": "OPENROUTER_API_KEY",
      "default_model": "anthropic/claude-sonnet-4.5"
    }
  ]
}
```

Copy or merge the generated providers into `ti.json` after review.

## Pack Manager

Generated packs are not blindly executed. A pack can be disabled by creating `.disabled` in the pack directory. This keeps risky imported packs reviewable.

## Safety default

External plugin/workflow execution should go through:

```text
translate -> ti.pack.yaml -> pack review -> sandbox policy -> audit/block logs
```

Do not run unknown Node/Python/shell plugins natively inside the Ti process.
