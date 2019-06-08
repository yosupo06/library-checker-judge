#!/usr/bin/env sh

cgexec -g cpuset,memory:/lib-judge capsh --chroot=sand --drop=cap_sys_chroot -- -c "$1"
