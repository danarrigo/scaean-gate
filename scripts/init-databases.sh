#!/bin/sh
set -eu

for database in sso_db app_a_db app_b_db; do
  if ! psql --username "$POSTGRES_USER" --dbname postgres --tuples-only --no-align \
    --command "SELECT 1 FROM pg_database WHERE datname = '$database'" | grep -q '^1$'; then
    createdb --username "$POSTGRES_USER" "$database"
  fi
done
