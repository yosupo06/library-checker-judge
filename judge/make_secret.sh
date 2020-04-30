#!/usr/bin/env bash

cat << EOF > secret.toml
postgre_host = "${PG_HOST:-localhost}"
postgre_user = "postgres"
postgre_pass = "${PG_PASS:-passwd}"
api_host = "${API_HOST:-localhost:50051}"
api_user = "judge"
api_pass = "${API_PASS:-password}"
${PROD:+prod="true"}
EOF
