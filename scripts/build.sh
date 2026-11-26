#!/usr/bin/env bash
# Testet und baut für Linux und Windows.
set -euo pipefail
cd "$(dirname "$0")/.."
go vet ./...
go test ./...
./scripts/build-linux.sh
./scripts/build-windows.sh
