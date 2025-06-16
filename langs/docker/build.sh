#!/usr/bin/env bash

set -e

SCRIPT_DIR=$(cd $(dirname $0); pwd)

docker build -t library-checker-images-gcc -f $SCRIPT_DIR/images/GCC.dockerfile $SCRIPT_DIR
docker build -t library-checker-images-ldc -f $SCRIPT_DIR/images/LDC.dockerfile $SCRIPT_DIR
docker build -t library-checker-images-python3 -f $SCRIPT_DIR/images/PYTHON3.dockerfile $SCRIPT_DIR
docker build -t library-checker-images-haskell -f $SCRIPT_DIR/images/HASKELL.dockerfile $SCRIPT_DIR
docker build -t library-checker-images-csharp -f $SCRIPT_DIR/images/CSHARP.dockerfile $SCRIPT_DIR
docker build -t library-checker-images-rust -f $SCRIPT_DIR/images/RUST.dockerfile $SCRIPT_DIR
docker build -t library-checker-images-java -f $SCRIPT_DIR/images/JAVA.dockerfile $SCRIPT_DIR
docker build -t library-checker-images-pypy -f $SCRIPT_DIR/images/PYPY.dockerfile $SCRIPT_DIR
docker build -t library-checker-images-golang -f $SCRIPT_DIR/images/GOLANG.dockerfile $SCRIPT_DIR
docker build -t library-checker-images-lisp -f $SCRIPT_DIR/images/LISP.dockerfile $SCRIPT_DIR
docker build -t library-checker-images-crystal -f $SCRIPT_DIR/images/CRYSTAL.dockerfile $SCRIPT_DIR
docker build -t library-checker-images-ruby -f $SCRIPT_DIR/images/RUBY.dockerfile $SCRIPT_DIR
