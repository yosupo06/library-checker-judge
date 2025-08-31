#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR=$(cd "$(dirname "$0")"; pwd)
exec python3 "$SCRIPT_DIR/build.py" "$@"
