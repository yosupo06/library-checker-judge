#!/usr/bin/env bash

set -e

apt-get install -y make git gcc build-essential pkgconf libtool \
   libsystemd-dev libprotobuf-c-dev libcap-dev libseccomp-dev libyajl-dev \
   go-md2man autoconf python3 automake

cd /root/
git clone https://github.com/containers/crun -b 1.4.5
cd /root/crun

./autogen.sh
./configure
make
