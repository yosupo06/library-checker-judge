#!/usr/bin/env sh

# prepare cgroups
cgcreate -g cpu,memory:/lib-judge
cgset -r memory.limit_in_bytes=1G lib-judge
cgset -r memory.memsw.limit_in_bytes=1G lib-judge

# prepare sandbox
rm -rf work
mkdir work

umount sand/proc
umount sand/dev
umount sand/sys
umount sand/bin
umount sand/lib
umount sand/lib64
umount sand/usr

rm -rf sand

mkdir sand sand/proc sand/dev sand/sys sand/bin sand/lib sand/lib64 sand/usr

mount -o ro -t proc none sand/proc
mount -o ro --bind /dev sand/dev
mount -o ro --bind /sys sand/sys
mount -o ro --bind /bin sand/bin
mount -o ro --bind /lib sand/lib
mount -o ro --bind /lib64 sand/lib64
mount -o ro --bind /usr sand/usr
