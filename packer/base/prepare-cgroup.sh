#!/usr/bin/env bash

set -e

# we have to launch docker once and re-set cpusets
# ref: https://github.com/yosupo06/library-checker-judge/issues/346
docker run --cgroup-parent judge.slice hello-world && true
echo '0,1' > /sys/fs/cgroup/judge.slice/cpuset.cpus
