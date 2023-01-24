#!/usr/bin/env bash

set -e

mkdir -p /var/lib/docker
mount -t tmpfs -o size=13g /dev/shm /var/lib/docker

rsync -aXS /var/lib/docker-base/. /var/lib/docker

mkdir /sys/fs/cgroup/judge-docker.slice/
echo '0,1' > /sys/fs/cgroup/test.slice/cpuset.cpus
echo 'root' > /sys/fs/cgroup/test.slice/cpuset.cpus.partition
