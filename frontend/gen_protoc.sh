#!/usr/bin/env bash

echo '@test'
ls $(pwd)

echo '@test2'
ls ../library-checker-judge/api/proto

cd .. && docker run -v $(pwd):/defs namely/protoc-all -i library-checker-judge/api/proto -f library_checker.proto -l web -o library-checker-frontend/src/api