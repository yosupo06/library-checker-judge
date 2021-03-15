#!/usr/bin/env bash
pdir=$(cd $(dirname $0)/..;pwd)
echo $pdir
docker run -v $pdir:/defs namely/protoc-all -i library-checker-judge/api/proto -f library_checker.proto -l web -o library-checker-frontend/src/api