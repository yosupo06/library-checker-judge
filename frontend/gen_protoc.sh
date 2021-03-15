#!/usr/bin/env bash
cp $1 .
docker run -v `pwd`:/defs namely/protoc-all -f library_checker.proto -l web -o src/api
