#!/usr/bin/env bash

set -e

# gcloud compute instances delete lib-judge2
gcloud compute instances create lib-judge2 --zone=asia-northeast1-c \
--machine-type=c2-standard-4 \
--metadata-from-file user-data=cloudinit.yml \
--image-family=ubuntu-1804-lts --image-project=ubuntu-os-cloud

until gcloud compute ssh root@lib-judge2 -- ls /root/can_start > /dev/null; do
    echo 'waiting...'
    sleep 10
done

echo "Build judge"
gcloud compute ssh root@lib-judge2 -- "
cd /root/library-checker-judge/judge &&
go build .
"

echo "Make Secret HOST=${PG_HOST} / PASS=${PG_PASS}"
gcloud compute ssh root@lib-judge2 -- "
cd /root/library-checker-judge/judge &&
PG_HOST=${PG_HOST} PG_PASS=${PG_PASS} ./make_secret.sh
"

echo "Launch Judge"
gcloud compute ssh root@lib-judge2 -- "supervisorctl start judge"
