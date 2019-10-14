#!/usr/bin/env bash

set -e

# gcloud compute instances delete lib-judge
gcloud compute instances create lib-judge-test --zone=asia-northeast1-c \
--machine-type=c2-standard-4 \
--boot-disk-size=200GB \
--metadata-from-file user-data=cloudinit.yml \
--image-family=ubuntu-1804-lts --image-project=ubuntu-os-cloud \
--preemptible

until gcloud compute ssh root@lib-judge-test -- ls /root/can_start > /dev/null; do
    echo 'waiting...'
    sleep 10
done

echo "Make Secret"
gcloud compute ssh root@lib-judge-test -- "cd /root/library-checker-judge/judge && ./make_secret.sh"

echo 'Start generate.py test'
gcloud compute ssh root@lib-judge-test -- "ulimit -s unlimited && cd /root/library-checker-problems && ./generate.py problems.toml"

echo 'Start executor.py test'
gcloud compute ssh root@lib-judge-test -- "cd /root/library-checker-judge/judge && ./executor_test.py"

echo 'Start docker test'
gcloud compute ssh root@lib-judge-test -- "cd /root/library-checker-judge/local && ./launch.sh"

echo 'Start judge test'
gcloud compute ssh root@lib-judge-test -- "cd /root/library-checker-judge/judge && go test . -v"

echo 'finish'
gcloud compute instances delete lib-judge-test --zone=asia-northeast1-c --quiet
