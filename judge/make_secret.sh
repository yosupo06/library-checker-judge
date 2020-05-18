#!/usr/bin/env bash

cat << EOF > secret.toml
api_host = "${API_HOST:-localhost:50051}"
api_user = "judge"
api_pass = "${API_PASS:-password}"
minio_host = "${MINIO_HOST:-localhost:9000}"
minio_access = "${MINIO_ACCESS:-minio}"
minio_secret = "${MINIO_SECRET:-miniopass}"
minio_bucket = "${MINIO_BUCKET:-testcase}"
${PROD:+prod="true"}
EOF
