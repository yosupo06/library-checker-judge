#!/usr/bin/env bash

set -ev

ENV=$(curl -X GET -H "Metadata-Flavor: Google" "http://metadata.google.internal/computeMetadata/v1/instance/attributes/env")

MINIO_HOST=$(gcloud secrets versions access latest --secret=minio-host)
MINIO_ID=$(gcloud secrets versions access latest --secret=minio-id)
MINIO_KEY=$(gcloud secrets versions access latest --secret=minio-secret)
MINIO_BUCKET=$(gcloud secrets versions access latest --secret=minio-bucket)

API_HOST=$(gcloud secrets versions access latest --secret=$ENV-api-host)
API_PASS=$(gcloud secrets versions access latest --secret=$ENV-api-judge-pass)

/root/judge \
-langs=/root/langs.toml \
-miniohost=$MINIO_HOST \
-minioid=$MINIO_ID \
-miniokey=$MINIO_KEY \
-miniobucket=$MINIO_BUCKET \
-apihost=$API_HOST \
-apipass=$API_PASS \
-cgroup-parent=judge.slice \
-prod