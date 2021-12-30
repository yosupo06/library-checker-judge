#!/bin/bash

mkdir -p /ramdisk
mount -t tmpfs -o size=4g /dev/shm /ramdisk
mkdir -p /compiler
mount -t tmpfs -o size=4g /dev/shm /compiler
