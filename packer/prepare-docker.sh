#!/usr/bin/env bash

set -e

service docker stop

mv /var/lib/docker /var/lib/docker-base

mkdir -p /var/lib/docker
mount -t tmpfs -o size=13g /dev/shm /var/lib/docker

rsync -aXS /var/lib/docker-base/. /var/lib/docker
rm -rf /var/lib/docker-base

service docker start
