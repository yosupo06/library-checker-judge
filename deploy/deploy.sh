#!/usr/bin/env bash

set -e

NAME=lib-judge-executor-$(cat /dev/urandom | LC_CTYPE=C tr -d -c '[:lower:]' | fold -w 10 | head -n 1)

./create_instance.sh $NAME $ZONE prod $CREATE_OPTION

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

#cd ../judge && CGO_ENABLED=0 GOOS=linux go build ../judge
#cd -
#gcloud compute scp ../judge/judge root@${NAME}:/root/judge --zone=${ZONE}
gcpexec "cp /root/library-checker-judge/judge/judge /root/judge"

gcloud compute scp ../langs/langs.toml root@${NAME}:/root/langs.toml --zone=${ZONE}
gcpexec "cp /root/library-checker-judge/judge/secret.toml /root/secret.toml"

gcpexec "mkdir -p /usr/local/lib/systemd/system/"
gcloud compute scp judge.service root@${NAME}:/usr/local/lib/systemd/system/judge.service --zone=${ZONE}

gcpexec "systemctl daemon-reload"
gcpexec "service judge start"
