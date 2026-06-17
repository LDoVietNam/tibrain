# Ti Token Optimizer

Ti Token Optimizer is a Go-native command-output compression layer inspired by rtk-style CLI proxies. It reduces context pollution before command output is passed to an LLM.

It is deterministic and local-first:

- no network calls
- no model required
- sandbox-aware command execution
- block recording
- cache for repeated outputs
- JSON output for automation

## Commands

```bash
ti-cli token estimate file.log
tail -n 1000 test.log | ti-cli token estimate -

ti-cli token compress test.log --command "go test ./..."
ti-cli token compress diff.txt --command "git diff" --max-lines 120

ti-cli token run -- go test ./...
ti-cli token run --backend portable --timeout 60 -- npm test

ti-cli proxy run -- git diff

ti-cli token cache path
ti-cli token cache stats
ti-cli token cache clear
```

## Compression modes

Ti detects common dev command outputs:

- `git-diff`: keeps file headers, hunks, and representative changed lines.
- `test-log`: keeps failures, assertions, stack traces, and tail summary.
- `install-log`: keeps warnings/errors/audit lines and tail summary.
- `json-log`: generic head/tail compression with error-line retention.
- `generic`: keeps head, tail, important lines, and collapses repeated output.

## Safety

`token run` executes commands through Ti sandbox. It records a Ti block with metadata:

- original estimated tokens
- compressed estimated tokens
- saved tokens
- compression kind
- sandbox id/backend

The optimizer does not claim exact tokenizer parity. Token counts are practical estimates for comparing before/after output size.
