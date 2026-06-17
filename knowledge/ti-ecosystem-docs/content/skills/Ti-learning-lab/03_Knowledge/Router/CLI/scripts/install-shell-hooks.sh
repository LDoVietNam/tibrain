#!/usr/bin/env sh
set -eu

SHELL_NAME="${1:-}"
if [ -z "$SHELL_NAME" ]; then
  SHELL_NAME="$(basename "${SHELL:-sh}")"
fi

case "$SHELL_NAME" in
  bash)
    TARGET="$HOME/.bashrc"
    SNIPPET='eval "$(ti-cli completion bash)"'
    ;;
  zsh)
    TARGET="$HOME/.zshrc"
    SNIPPET='source <(ti-cli completion zsh)'
    ;;
  fish)
    TARGET="$HOME/.config/fish/config.fish"
    mkdir -p "$(dirname "$TARGET")"
    SNIPPET='ti-cli completion fish | source'
    ;;
  *)
    echo "unsupported shell: $SHELL_NAME" >&2
    exit 1
    ;;
esac

if [ ! -f "$TARGET" ] || ! grep -Fq "$SNIPPET" "$TARGET" 2>/dev/null; then
  printf '\n# Ti CLI completion\n%s\n' "$SNIPPET" >> "$TARGET"
  echo "installed Ti completion in $TARGET"
else
  echo "Ti completion already installed in $TARGET"
fi
