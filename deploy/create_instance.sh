#!/usr/bin/env bash

# ./create_instance.sh NAME ZONE ARG

set -e

NAME=$1
ZONE=$2

echo "Create Instance Name = $NAME, Zone = $ZONE, Extra Opt = ${@:3}"
gcloud compute instances create $NAME --zone=$ZONE \
--machine-type=c2-standard-4 \
--boot-disk-size=50GB \
--boot-disk-type=pd-ssd \
--metadata-from-file user-data=cloudinit.yml \
--image-family=judge-image-family ${@:3}

function gcpexec() {
    echo "Start: ${1}"
    gcloud compute ssh root@${NAME} --zone ${ZONE} -- $1
    RET=$?
    echo "Finish: ${1}"
    return $RET
}

until gcpexec "ls /root/can_start > /dev/null"; do
    echo 'waiting...'
    sleep 10
done

echo "Make judge.tar.gz(compressed library-checker-judge)"
if [ -d "../.git" ]; then
    files=$(cd .. && git ls-files)
else
    files=$(cd .. && find .)
fi
(cd .. && tar -cf judge.tar.gz $files)

echo "Copy judge.tar.gz"
gcpexec "cd /root/ && mkdir library-checker-judge"
gcloud compute scp --zone ${ZONE} ../judge.tar.gz root@${NAME}:/root/library-checker-judge/judge.tar.gz
echo "Extract judge.tar.gx"
gcpexec "cd /root/library-checker-judge && tar -xf judge.tar.gz"
echo "Generate protoc"
gcpexec "cd /root/library-checker-judge && ./api/gen_protoc.sh"

echo "Install executor"
gcpexec "cd /root/library-checker-judge/judge/executor_rust && cargo install --path . --root /usr/ --features sandbox"

echo "Build judge"
gcpexec "cd /root/library-checker-judge/judge && go build ."
