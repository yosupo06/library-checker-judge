#!/usr/bin/env bash

cat << EOF > secret.toml
postgre_host = "${PG_HOST:-localhost}"
postgre_user = "postgres"
postgre_pass = "${PG_PASS:-passwd}"
api_host = "${API_HOST:-localhost:50051}"
api_user = "judge"
api_pass = "${API_PASS:-password}"
minio_host = "${MINIO_HOST:-localhost:9000}"
minio_access = "${MINIO_ACCESS:-minio}"
minio_secret = "${MINIO_SECRET:-miniopass}"
minio_bucket = "${MINIO_BUCKET:-testcase}"
${PROD:+prod="true"}
EOF
