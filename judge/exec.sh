#!/usr/bin/env sh
cgexec -g memory:/lib-judge chroot sand $1
#chroot sand cgexec -g memory:/lib-judge $1