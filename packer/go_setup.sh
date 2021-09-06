#!/usr/bin/env bash

# install go1.17

wget -q https://golang.org/dl/go1.17.linux-amd64.tar.gz -O /tmp/go1.17.linux-amd64.tar.gz
sudo tar -C /opt/ -xzf /tmp/go1.17.linux-amd64.tar.gz

sudo ln -s /opt/go/bin/go /usr/bin/
