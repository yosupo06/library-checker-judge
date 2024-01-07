#!/usr/bin/env bash

set -ev

/root/judge \
-langs=/root/langs.toml \
-miniohost=$MINIO_HOST \
-minioid=$MINIO_ID \
-miniokey=$MINIO_KEY \
-miniobucket=$MINIO_BUCKET \
-miniopublicbucket=$MINIO_PUBLIC_BUCKET \
-pguser=$PG_USER \
-prod
