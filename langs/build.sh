#!/usr/bin/env bash

set -e

SCRIPT_DIR=$(cd $(dirname $0); pwd)

docker build -t library-checker-images-gcc -f $SCRIPT_DIR/Dockerfile.GCC $SCRIPT_DIR
docker build -t library-checker-images-ldc -f $SCRIPT_DIR/Dockerfile.LDC $SCRIPT_DIR
docker build -t library-checker-images-python3 -f $SCRIPT_DIR/Dockerfile.PYTHON3 $SCRIPT_DIR
docker build -t library-checker-images-haskell -f $SCRIPT_DIR/Dockerfile.HASKELL $SCRIPT_DIR
docker build -t library-checker-images-csharp -f $SCRIPT_DIR/Dockerfile.CSHARP $SCRIPT_DIR
docker build -t library-checker-images-rust -f $SCRIPT_DIR/Dockerfile.RUST $SCRIPT_DIR
docker build -t library-checker-images-java -f $SCRIPT_DIR/Dockerfile.JAVA $SCRIPT_DIR
docker build -t library-checker-images-pypy -f $SCRIPT_DIR/Dockerfile.PYPY $SCRIPT_DIR
docker build -t library-checker-images-golang -f $SCRIPT_DIR/Dockerfile.GOLANG $SCRIPT_DIR
docker build -t library-checker-images-lisp -f $SCRIPT_DIR/Dockerfile.LISP $SCRIPT_DIR
docker build -t library-checker-images-crystal -f $SCRIPT_DIR/Dockerfile.CRYSTAL $SCRIPT_DIR
docker build -t library-checker-images-ruby -f $SCRIPT_DIR/Dockerfile.RUBY $SCRIPT_DIR
docker build -t library-checker-images-swift -f $SCRIPT_DIR/Dockerfile.SWIFT $SCRIPT_DIR