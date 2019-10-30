cat << EOF > secret.yaml
env_variables:
  POSTGRE_HOST: $PG_HOST
  POSTGRE_USER: postgres
  POSTGRE_PASS: $PG_PASS
  SESSION_SECRET: $SESSION_SECRET
EOF
#TODO
apt-get update
apt-get -y install wget
wget https://raw.githubusercontent.com/yosupo06/library-checker-judge/master/compiler/langs.toml