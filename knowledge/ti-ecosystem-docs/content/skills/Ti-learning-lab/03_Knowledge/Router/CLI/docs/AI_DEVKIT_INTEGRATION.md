# AI DevKit Integration

This pack includes a Go-native subset of the uploaded AI DevKit workflow:

- phase templates from `ai-devkit`;
- reusable command prompts;
- project initialization;
- project readiness linting;
- prompt execution through Ti providers.

## Initialize a project

```bash
ti-cli devkit init --all
```

Default output:

```text
.ai-devkit.json
docs/ai/requirements.md
docs/ai/design.md
docs/ai/planning.md
docs/ai/implementation.md
docs/ai/testing.md
docs/ai/deployment.md
docs/ai/monitoring.md
```

## Add or inspect phases

```bash
ti-cli devkit phase
ti-cli devkit phase show requirements
ti-cli devkit phase implementation --overwrite
```

## Reusable prompts

```bash
ti-cli devkit command
ti-cli devkit command show code-review
```

You can also run them directly:

```bash
git diff | ti-cli prompt run code-review --provider openrouter
```

## Lint readiness

```bash
ti-cli devkit lint
ti-cli devkit lint --json
```
