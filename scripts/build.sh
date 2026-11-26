#!/usr/bin/env bash
# Tests and builds for Linux and Windows.
set -euo pipefail
cd "$(dirname "$0")/.."
go vet ./...
go test ./...
./scripts/build-linux.sh
./scripts/build-windows.sh
