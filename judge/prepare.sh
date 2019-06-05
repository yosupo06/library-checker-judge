#!/usr/bin/env sh

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
mount -o ro,bind /dev sand/dev
mount -o ro,bind /sys sand/sys
mount -o ro,bind /bin sand/bin
mount -o ro,bind /lib sand/lib
mount -o ro,bind /lib64 sand/lib64
mount -o ro,bind /usr sand/usr
