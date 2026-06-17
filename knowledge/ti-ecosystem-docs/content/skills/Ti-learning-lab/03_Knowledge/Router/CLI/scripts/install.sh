#!/usr/bin/env bash
set -euo pipefail

PREFIX="${PREFIX:-$HOME/.local}"
make build
mkdir -p "$PREFIX/bin"
cp "bin/ti-cli" "$PREFIX/bin/ti-cli"
echo "Installed ti-cli to $PREFIX/bin/ti-cli"
echo "Add this to PATH if needed: export PATH=\"$PREFIX/bin:\$PATH\""
