#!/usr/bin/env sh

./api -hmackey=$HMAC_KEY -pghost=$PG_HOST -pgtable=$PG_TABLE -pguser=$PG_USER -pgpass=$PG_PASS $@