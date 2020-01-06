#!/usr/bin/env bash

set -e

NAME=lib-judge-test-$(cat /dev/urandom | LC_CTYPE=C tr -d -c '[:lower:]' | fold -w 10 | head -n 1)
ZONE=asia-northeast1-c

if [ $# -ge 1 ] && [ $1 = "remain" ]; then
    echo "[WARN!] Remain Instance"
else
    echo "Auto Release"
    trap "echo 'Release' && gcloud compute instances delete ${NAME} --zone=${ZONE} --quiet" 0
fi

(cd ../deploy && ./create_instance.sh $NAME $ZONE)

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
gcpexec "cd /root/library-checker-problems && pip3 install -r requirements.txt"

# echo "Copy library-checker-judge : $(cd .. && pwd)"
# gcloud compute scp --zone ${ZONE} --recurse $(cd .. && pwd) root@${NAME}:/root/library-checker-judge

# echo "Install pip"
# gcpexec "pip3 install psutil psycopg2"

echo "Make Secret"
gcpexec "cd /root/library-checker-judge/judge && ./make_secret.sh"

# echo "Install compilers"
# gcpexec "cd /root/library-checker-judge/deploy && ./install.sh"

echo 'Start executor.py test'
gcpexec "cd /root/library-checker-judge/judge/executor && ./executor_test.py"

echo 'Start generate.py test'
gcpexec "ulimit -s unlimited && cd /root/library-checker-problems && ./generate.py problems.toml"

echo 'Start docker test'
gcpexec "cd /root/library-checker-judge && ./launch_local.sh"

echo 'Start judge test'
gcpexec "cd /root/library-checker-judge/judge && go test . -v"

echo 'Start API test'
gcpexec "cd /root/library-checker-judge/api && go test ."

