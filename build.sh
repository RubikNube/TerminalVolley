#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

mkdir -p bin

go build -o bin/terminalvolley ./cmd/terminalvolley

echo "Built: $ROOT_DIR/bin/terminalvolley"
