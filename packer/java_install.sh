#!/usr/bin/env bash

sudo apt-get install -y openjdk-17-jdk
sudo /var/lib/cloud/scripts/per-boot/10_java_setup.sh

sudo update-alternatives --install /usr/bin/java java /ramdisk/compiler/java-17-openjdk-amd64/bin/java 100000
sudo update-alternatives --install /usr/bin/javac javac /ramdisk/compiler/java-17-openjdk-amd64/bin/javac 100000
