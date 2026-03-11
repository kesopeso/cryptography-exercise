#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

if ! command -v docker &>/dev/null; then
    echo "Error: docker is not installed." >&2
    echo "Install it from: https://docs.docker.com/get-docker/" >&2
    exit 1
fi

if [[ $# -eq 0 ]]; then
    echo "Usage: $0 [-database <connection_string>] <command> [args]" >&2
    echo "" >&2
    echo "Commands:" >&2
    echo "  up [N]       Apply all or N up migrations" >&2
    echo "  down [N]     Apply all or N down migrations" >&2
    echo "  goto V       Migrate to version V" >&2
    echo "  force V      Set version V (no migration run)" >&2
    echo "  version      Print current migration version" >&2
    echo "  drop [-f]    Drop everything in the database" >&2
    echo "" >&2
    echo "If -database is not provided, the default local connection string is used." >&2
    exit 1
fi

DEFAULT_DATABASE="postgresql://postgres:postgres@localhost:5432/apidb?sslmode=disable"

# Check if -database flag is provided
HAS_DATABASE=false
for arg in "$@"; do
    if [[ "$arg" == "-database" || "$arg" == "-database="* ]]; then
        HAS_DATABASE=true
        break
    fi
done

if [[ "$HAS_DATABASE" == false ]]; then
    docker run --rm -it --network host -v "$(pwd)/migrations:/migrations" migrate/migrate \
        -path /migrations -database "$DEFAULT_DATABASE" "$@"
else
    docker run --rm -it --network host -v "$(pwd)/migrations:/migrations" migrate/migrate \
        -path /migrations "$@"
fi
