#!/usr/bin/env bash
cd .. && docker run -v `pwd`:/defs namely/protoc-all -i library-checker-judge/api/proto -f library_checker.proto -l web -o library-checker-frontend/src/api