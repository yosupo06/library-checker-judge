#!/usr/bin/env bash

set -e

wget https://github.com/containers/crun/releases/download/1.15/crun-1.15-linux-amd64 -O /root/crun
chmod +x /root/crun
