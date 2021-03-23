#!/usr/bin/env bash

set -e -o pipefail

SCRIPT_DIR=$(cd $(dirname $0); pwd)
docker run -v $SCRIPT_DIR:/defs namely/protoc-all:1.34_4 -f proto/library_checker.proto -l go -o .
