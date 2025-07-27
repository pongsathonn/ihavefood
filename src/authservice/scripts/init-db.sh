#!/bin/bash
set -e

# postgres image required $POSTGRES_PASSWORD for superuser password.
# $POSTGRES_USER is optional (default is postgres)
echo "$POSTGRES_USER"

AUTH_USER="${AUTH_DB_USER}"
AUTH_PASS="${AUTH_DB_PASS}"
AUTH_DB="${AUTH_DB_NAME}"

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER"  <<-EOSQL
    CREATE USER $AUTH_USER WITH PASSWORD '$AUTH_PASS';
    CREATE DATABASE $AUTH_DB;
    GRANT ALL PRIVILEGES ON DATABASE $AUTH_DB TO $AUTH_USER;
EOSQL

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$AUTH_DB" \
    -a -f /sql/create_table.sql

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$AUTH_DB" <<-EOSQL
    GRANT SELECT, INSERT, UPDATE, DELETE ON credentials TO $AUTH_USER;
EOSQL

