#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

go build -o ./server ./cmd/server
echo "server built successfully"
