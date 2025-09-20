#!/usr/bin/env bash

set -e

mkdir -p /var/lib/docker
mount -t tmpfs -o size=20g /dev/shm /var/lib/docker

rsync -aXS /var/lib/docker-base/. /var/lib/docker

mkdir /sys/fs/cgroup/judge.slice/


echo '0' > /sys/fs/cgroup/judge.slice/cpuset.cpus
echo 'isolated' > /sys/fs/cgroup/judge.slice/cpuset.cpus.partition
