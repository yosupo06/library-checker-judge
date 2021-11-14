#!/usr/bin/env bash

set -e -o pipefail

cp $1 .
docker run -v `pwd`:/defs namely/protoc-all:1.39_0 -f library_checker.proto -l web -o src/api
