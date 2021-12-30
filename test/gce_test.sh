#!/usr/bin/env bash

set -e

ENV=$1

echo "Start Test env=${ENV}"

NAME=lib-judge-test-$(date +%d-%s)-$RANDOM
ZONE=asia-east1-c

if [ $# -ge 2 ] && [ $2 = "remain" ]; then
    echo "[WARN!] Remain Instance"
else
    echo "Auto Release"
    trap "echo 'Release' && gcloud compute instances delete ${NAME} --zone=${ZONE} --quiet" 0
fi

(cd ../deploy && ./create_instance.sh $NAME $ZONE $ENV --preemptible)

function gcpexec() {
    echo "Start: ${1}"
    gcloud compute ssh root@${NAME} --zone ${ZONE} -- $1
    RET=$?
    echo "Finish: ${1}"
    return $RET
}

echo "Make problems.tar.gz(compressed library-checker-problems)"
(cd ../../library-checker-problems && tar -cf problems.tar.gz $(git ls-files))
echo "Copy problems.tar.gz"
gcpexec "cd /root/ && mkdir library-checker-problems"
gcloud compute scp --zone ${ZONE} ../../library-checker-problems/problems.tar.gz root@${NAME}:/root/library-checker-problems/problems.tar.gz
echo "Extract problems.tar.gx"
gcpexec "cd /root/library-checker-problems && tar -xf problems.tar.gz"

echo "Install pip"
gcpexec "cd /root/library-checker-problems && python3 -m pip install -r requirements.txt"

echo "Make Secret"
gcpexec "cd /root/library-checker-judge/judge && ./make_secret.sh"

echo 'Start executor_rust test'
gcpexec "cd /root/library-checker-judge/executor && cargo test -- --test-threads=1"

echo 'Start generate.py test'
gcpexec "ulimit -s unlimited && cd /root/library-checker-problems && ./generate.py -p aplusb unionfind"

echo 'Start docker test'
gcpexec "cd /root/library-checker-judge && ./launch_local.sh"

echo 'Start judge test'
gcpexec "cd /root/library-checker-judge/judge && go test . -v"

echo 'Start API test'
gcpexec "cd /root/library-checker-judge/api && go test ."

