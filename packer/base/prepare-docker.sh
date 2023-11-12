#!/usr/bin/env bash

set -e

mkdir -p /var/lib/docker
mount -t tmpfs -o size=13g /dev/shm /var/lib/docker

rsync -aXS /var/lib/docker-base/. /var/lib/docker

mkdir /sys/fs/cgroup/judge.slice/


echo '0,1' > /sys/fs/cgroup/judge.slice/cpuset.cpus
echo 'isolated' > /sys/fs/cgroup/judge.slice/cpuset.cpus.partition

# we have to launch docker once and re-set cpusets
# ref: https://github.com/yosupo06/library-checker-judge/issues/346
docker run --cgroup-parent judge.slice hello-world && true
echo '0,1' > /sys/fs/cgroup/judge.slice/cpuset.cpus
