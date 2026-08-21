#!/bin/sh
set -eu

for database in sso_db app_a_db app_b_db; do
  psql --username "$POSTGRES_USER" --dbname postgres --set ON_ERROR_STOP=1 \
    --command "SELECT 'CREATE DATABASE ' || quote_ident('$database') WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$database') \\gexec"
done
