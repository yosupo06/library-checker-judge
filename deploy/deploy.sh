#!/usr/bin/env bash

set -e

NAME=lib-judge-executor-$(cat /dev/urandom | LC_CTYPE=C tr -d -c '[:lower:]' | fold -w 10 | head -n 1)
ZONE=asia-northeast1-c

./create_instance.sh $NAME $ZONE

function gcpexec() {
    echo "Start: ${1}"
    gcloud compute ssh root@${NAME} --zone ${ZONE} -- $1
    RET=$?
    echo "Finish: ${1}"
    return $RET
}

echo "Make Secret HOST=${PG_HOST} / PASS=${PG_PASS}"
gcpexec "cd /root/library-checker-judge/judge && PG_HOST=${PG_HOST} PG_PASS=${PG_PASS} ./make_secret.sh"

gcpexec "supervisorctl restart judge"
