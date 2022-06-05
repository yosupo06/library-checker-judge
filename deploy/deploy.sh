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

#cd ../judge && CGO_ENABLED=0 GOOS=linux go build ../judge
#cd -
#gcloud compute scp ../judge/judge root@${NAME}:/root/judge --zone=${ZONE}
gcpexec "cp /root/library-checker-judge/judge/judge /root/judge"

gcloud compute scp ../langs/langs.toml root@${NAME}:/root/langs.toml --zone=${ZONE}

gcpexec "mkdir -p /usr/local/lib/systemd/system/"
gcloud compute scp judge.service root@${NAME}:/usr/local/lib/systemd/system/judge.service --zone=${ZONE}

gcpexec "systemctl daemon-reload"
gcpexec "service judge start"
