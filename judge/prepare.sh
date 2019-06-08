#!/usr/bin/env sh

# prepare cgroups
cgcreate -g cpuset,memory:/lib-judge
# String 1 Core
cgset -r cpuset.cpus=0 lib-judge
cgset -r cpuset.mems=0 lib-judge
# Memory limit is 1G
cgset -r memory.limit_in_bytes=1G lib-judge
cgset -r memory.memsw.limit_in_bytes=1G lib-judge

# prepare sandbox
rm -rf work
mkdir work

rm -rf sand
mkdir sand sand/proc sand/dev sand/sys sand/bin sand/lib sand/lib64 sand/usr

# mount -o ro -t proc none sand/proc
# mount -o ro --bind /dev sand/dev
# mount -o ro --bind /sys sand/sys
# mount -o ro --bind /bin sand/bin
# mount -o ro --bind /lib sand/lib
# mount -o ro --bind /lib64 sand/lib64
# mount -o ro --bind /usr sand/usr

mount -t proc none sand/proc
mount --bind /dev sand/dev
mount --bind /sys sand/sys
mount --bind /bin sand/bin
mount --bind /lib sand/lib
mount --bind /lib64 sand/lib64
mount --bind /usr sand/usr
