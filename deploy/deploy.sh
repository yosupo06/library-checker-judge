#!/usr/bin/env bash

set -e

cp cloudinit.yml cloudinit_buf.yml

sed -i -e "s/{POSTGRE_HOST}/$POSTGRE_HOST/" cloudinit_buf.yml
sed -i -e "s/{POSTGRE_PASS}/$POSTGRE_PASS/" cloudinit_buf.yml

# gcloud compute instances delete lib-judge
gcloud compute instances create lib-judge --zone=asia-northeast1-c \
--machine-type=n1-highcpu-2 \
--metadata-from-file user-data=cloudinit_buf.yml \
--image-family=ubuntu-1804-lts --image-project=ubuntu-os-cloud
