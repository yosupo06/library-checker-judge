#!/usr/bin/env bash

set -e -o pipefail

SCRIPT_DIR=$(cd $(dirname $0); pwd)

# compile api proto
docker run -u `id -u`:`id -g` -v $SCRIPT_DIR:/defs namely/protoc-all:1.51_2 -l go -f api/proto/library_checker.proto -o api --go-module-prefix github.com/yosupo06/library-checker-judge/api
