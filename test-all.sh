#!/bin/bash
set -e

echo "=== Testing Library Checker Judge Components ==="

# Launch local development environment
echo "Starting local development environment..."
if ! docker ps | grep -q postgres; then
    echo "PostgreSQL not running, starting local environment..."
    ./launch_local.sh &
    LAUNCH_PID=$!
    
    # Wait for PostgreSQL to be ready
    echo "Waiting for PostgreSQL to be ready..."
    timeout=60
    while ! docker ps | grep -q postgres && [ $timeout -gt 0 ]; do
        sleep 1
        ((timeout--))
    done
    
    if [ $timeout -eq 0 ]; then
        echo "Timeout waiting for PostgreSQL to start"
        exit 1
    fi
    
    # Additional wait for PostgreSQL to accept connections
    sleep 10
    echo "PostgreSQL is ready"
    CLEANUP_NEEDED=true
else
    echo "PostgreSQL already running"
    CLEANUP_NEEDED=false
fi

echo "Testing database..."
cd database && go test ./... -v && cd ..

echo "Testing API..."
cd api && go test ./... -v && cd ..

echo "Testing storage..."
cd storage && go test ./... -v && cd ..

echo "Testing uploader..."
if [[ -d "uploader" ]]; then
    cd uploader && go test ./... -v && cd ..
fi

echo "Building all components..."
go build ./api/...
go build ./database/...
go build ./storage/...

echo "Running static analysis..."
go vet ./...
gofmt -l . | (! read)

# Cleanup if we started the environment
if [ "$CLEANUP_NEEDED" = true ]; then
    echo "Cleaning up local environment..."
    docker compose down
fi

echo "=== All tests passed! ==="