#!/usr/bin/env bash

set -e -o pipefail

SCRIPT_DIR=$(cd $(dirname $0); pwd)
docker run -v $SCRIPT_DIR:/defs namely/protoc-all:1.37_2 -l go -f proto/library_checker.proto -o . --go-module-prefix github.com/yosupo06/library-checker-judge/api
