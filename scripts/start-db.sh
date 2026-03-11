#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

if [[ $# -ne 0 ]]; then
    echo "Usage: $0" >&2
    exit 1
fi

docker compose up -d
