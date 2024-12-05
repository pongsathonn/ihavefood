#!/bin/bash
set -e

echo "$POSTGRES_USER"

# fetch env variable from .env file
AUTH_USER="${AUTH_POSTGRES_USER}"
AUTH_PASS="${AUTH_POSTGRES_PASS}"
AUTH_DB="${AUTH_POSTGRES_DATABASE}"

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


