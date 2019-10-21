#!/usr/bin/env bash

set -e

NAME=lib-judge-test-$(cat /dev/urandom | tr -d -c '[:lower:]' | fold -w 10 | head -n 1)
ZONE=asia-northeast1-c


echo "Set default zone : ${ZONE}"
gcloud compute set compute/zone $ZONE

echo "Create ${NAME}"

gcloud compute instances create $NAME --zone=$ZONE \
--machine-type=c2-standard-4 \
--boot-disk-size=200GB \
--metadata-from-file user-data=cloudinit.yml \
--image-family=ubuntu-1804-lts --image-project=ubuntu-os-cloud \
--preemptible

trap "echo 'Release' && gcloud compute instances delete ${NAME} --zone=${ZONE} --quiet" 0

exit 0

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

