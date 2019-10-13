cat << EOF > secret.toml
postgre_host = "${PG_HOST:-localhost}"
postgre_user = "postgres"
postgre_pass = "${PG_PASS:-passwd}"
EOF
