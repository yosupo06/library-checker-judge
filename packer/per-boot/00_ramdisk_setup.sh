#!/bin/bash

mkdir -p /ramdisk
mount -t tmpfs -o size=8g /dev/shm /ramdisk
mkdir /ramdisk/compiler
