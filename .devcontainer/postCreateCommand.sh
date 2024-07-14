#!/usr/bin/env bash

set -e

sudo apt-get update
sudo apt-get install -y postgresql

git clone https://github.com/yosupo06/library-checker-problems $HOME/library-checker-problems
