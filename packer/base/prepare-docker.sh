#!/usr/bin/env bash

set -e

mkdir -p /var/lib/docker
mount -t tmpfs -o size=13g /dev/shm /var/lib/docker

rsync -aXS /var/lib/docker-base/. /var/lib/docker

mkdir /sys/fs/cgroup/judge.slice/


echo '0,1' > /sys/fs/cgroup/judge.slice/cpuset.cpus
# somehow it breaks docker
# echo 'root' > /sys/fs/cgroup/judge.slice/cpuset.cpus.partition
