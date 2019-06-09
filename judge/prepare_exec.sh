#!/usr/bin/env sh

cgdelete cpuset,memory:/lib-judge
# prepare cgroups
cgcreate -g cpuset,memory:/lib-judge
# String 1 Core
cgset -r cpuset.cpus=0 lib-judge
cgset -r cpuset.mems=0 lib-judge
# Memory limit is 1G
cgset -r memory.limit_in_bytes=1G lib-judge
cgset -r memory.memsw.limit_in_bytes=1G lib-judge
