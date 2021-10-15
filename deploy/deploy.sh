#!/usr/bin/env bash

set -e

NAME=lib-judge-executor-$(cat /dev/urandom | LC_CTYPE=C tr -d -c '[:lower:]' | fold -w 10 | head -n 1)

./create_instance.sh $NAME $ZONE $CREATE_OPTION

function gcpexec() {
    echo "Start: ${1}"
    gcloud compute ssh root@${NAME} --zone ${ZONE} -- $1
    RET=$?
    echo "Finish: ${1}"
    return $RET
}

echo "Make Secret HOST=${PG_HOST} / PASS=${PG_PASS}"
gcpexec "cd /root/library-checker-judge/judge &&
    API_HOST=apiv1.yosupo.jp:443
    API_PASS=${API_PASS}
    MINIO_HOST=${MINIO_HOST}
    MINIO_ACCESS=${MINIO_ACCESS}
    MINIO_SECRET=${MINIO_SECRET}
    MINIO_BUCKET=${MINIO_BUCKET}
    PROD=true
    ./make_secret.sh"
gcpexec "cp /etc/supervisor/conf.d/judge._conf /etc/supervisor/conf.d/judge.conf"
gcpexec "supervisorctl reload"
