#!/usr/bin/env bash

set -e -o pipefail

cp ../api/proto/library_checker.proto .
docker run -v `pwd`:/defs namely/protoc-all:1.37_2 -f library_checker.proto -l python -o .
rm library_checker.proto
