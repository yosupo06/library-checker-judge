#!/bin/bash
set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
cd "$SCRIPT_DIR"

CLEANUP_NEEDED=false

cleanup() {
    if [[ "$CLEANUP_NEEDED" == true ]]; then
        echo "Cleaning up local environment..."
        docker compose down
    fi
}
trap cleanup EXIT

is_local_environment_running() {
    docker compose ps --services --status running | grep -qx "db" &&
        docker compose ps --services --status running | grep -qx "api-rest" &&
        docker compose ps --services --status running | grep -qx "gcs"
}

echo "=== Testing Library Checker Judge Components ==="

# Launch local development environment
echo "Starting local development environment..."
if ! is_local_environment_running; then
    echo "Local compose environment not running, starting it..."
    CLEANUP_NEEDED=true
    ./launch_local.sh
else
    echo "Local compose environment already running"
fi

echo "Testing database..."
(cd database && go test ./... -v)

echo "Testing REST API..."
(cd restapi && go test ./... -v)

echo "Testing storage..."
(cd storage && go test ./... -v)

echo "Testing uploader..."
if [[ -d "uploader" ]]; then
    (cd uploader && go test ./... -v)
fi

echo "Building all components..."
for module in restapi database storage; do
    (cd "$module" && go build ./...)
done

echo "Running static analysis..."
for module in restapi database storage uploader judge executor integration langs migrator utils cloudrun/taskqueue-metrics; do
    if [[ -d "$module" ]]; then
        (cd "$module" && go vet ./...)
    fi
done
gofmt -l . | (! read)

echo "=== All tests passed! ==="
