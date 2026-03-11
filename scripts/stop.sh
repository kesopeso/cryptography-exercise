#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

if [[ $# -eq 0 ]]; then
    docker compose down
elif [[ $# -eq 1 && "$1" == "--remove-data-volume" ]]; then
    docker compose down -v
else
    echo "Usage: $0 [--remove-data-volume]" >&2
    exit 1
fi
