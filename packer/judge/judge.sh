#!/usr/bin/env bash

set -ev

ENV=$(curl -X GET -H "Metadata-Flavor: Google" "http://metadata.google.internal/computeMetadata/v1/instance/attributes/env")

MINIO_HOST=$(gcloud secrets versions access latest --secret=minio-host)
MINIO_ID=$(gcloud secrets versions access latest --secret=minio-id)
MINIO_KEY=$(gcloud secrets versions access latest --secret=minio-secret)
MINIO_BUCKET=$(gcloud secrets versions access latest --secret=$ENV-minio-bucket)
MINIO_PUBLIC_BUCKET=$(gcloud secrets versions access latest --secret=$ENV-minio-public-bucket)

PG_HOST=$(gcloud secrets versions access latest --secret=pg-private-ip)
PG_USER=$(gcloud secrets versions access latest --secret=$ENV-pg-user)
PG_PASS=$(gcloud secrets versions access latest --secret=$ENV-pg-pass)
PG_TABLE=$(gcloud secrets versions access latest --secret=$ENV-pg-table)

/root/judge \
-langs=/root/langs.toml \
-miniohost=$MINIO_HOST \
-minioid=$MINIO_ID \
-miniokey=$MINIO_KEY \
-miniobucket=$MINIO_BUCKET \
-miniopublicbucket=$MINIO_PUBLIC_BUCKET \
-pghost=$PG_HOST \
-pguser=$PG_USER \
-pgpass=$PG_PASS \
-pgtable=$PG_TABLE \
-cgroup-parent=judge.slice \
-prod