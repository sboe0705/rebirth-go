#!/usr/bin/env bash
# Builds the Linux version (requires CGO + a C compiler, see README).
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p bin
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o bin/rebirth ./cmd/rebirth
echo "-> bin/rebirth"
