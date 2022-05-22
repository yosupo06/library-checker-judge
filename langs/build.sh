#!/usr/bin/env bash

set -e

SCRIPT_DIR=$(cd $(dirname $0); pwd)

docker build -t library-checker-images-gcc -f $SCRIPT_DIR/Dockerfile.GCC $SCRIPT_DIR
docker build -t library-checker-images-ldc -f $SCRIPT_DIR/Dockerfile.LDC $SCRIPT_DIR
docker build -t library-checker-images-python3 -f $SCRIPT_DIR/Dockerfile.PYTHON3 $SCRIPT_DIR
docker build -t library-checker-images-haskell -f $SCRIPT_DIR/Dockerfile.HASKELL $SCRIPT_DIR
docker build -t library-checker-images-csharp -f $SCRIPT_DIR/Dockerfile.CSHARP $SCRIPT_DIR

docker pull rust:1.60-slim
docker pull openjdk:17
docker pull pypy:3.9-7.3.9-slim
docker pull golang:1.18.2-alpine3.15
docker pull clfoundation/sbcl:2.1.5-slim

docker image prune -f
