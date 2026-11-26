#!/usr/bin/env bash
# Cross-Compile für Windows. Ebitengine braucht unter Windows kein CGO.
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p bin
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-s -w -H=windowsgui" -o bin/rebirth.exe ./cmd/rebirth
echo "-> bin/rebirth.exe"
