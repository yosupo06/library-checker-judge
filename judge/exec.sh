#!/usr/bin/env sh

unshare -n cgexec -g memory:/lib-judge capsh --chroot=sand --drop=cap_sys_chroot -- -c "$1"

#cgexec -g memory:/lib-judge chroot sand $1
#chroot sand cgexec -g memory:/lib-judge $1