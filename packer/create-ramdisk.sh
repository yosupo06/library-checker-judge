#!/usr/bin/env bash

set -e

mkdir -p /ramdisk
mount -t tmpfs -o size=1g /dev/shm /ramdisk
