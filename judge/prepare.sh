#!/usr/bin/env sh

# prepare sandbox
mkdir sand/proc sand/dev sand/sys sand/bin sand/lib sand/lib64 sand/usr

mount -t proc none sand/proc
mount --bind /dev sand/dev
mount --bind /sys sand/sys
mount --bind /bin sand/bin
mount --bind /lib sand/lib
mount --bind /lib64 sand/lib64
mount --bind /usr sand/usr
