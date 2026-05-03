#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

check_dockerfile() {
  local dockerfile="$1"
  local context="$2"

  echo "Checking ${dockerfile}"
  docker buildx build \
    --check \
    --build-arg 'BUILDKIT_DOCKERFILE_CHECK=error=true' \
    -f "${dockerfile}" \
    "${context}"
}

for dockerfile in Dockerfile.* firebase/Dockerfile; do
  check_dockerfile "${dockerfile}" .
done

for dockerfile in langs/Dockerfile.*; do
  check_dockerfile "${dockerfile}" langs
done
