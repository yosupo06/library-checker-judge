cat << EOF > secret.yaml
env_variables:
  POSTGRE_HOST: $PG_HOST
  POSTGRE_USER: postgres
  POSTGRE_PASS: $PG_PASS
  SESSION_SECRET: $SESSION_SECRET
EOF
#TODO
wget https://github.com/yosupo06/library-checker-judge/blob/master/compiler/langs.toml