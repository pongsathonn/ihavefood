#!/bin/bash
set -e

# this script will be executed in container

AUTH_USER="${AUTH_POSTGRES_USER}"
AUTH_PASSWORD="${AUTH_POSTGRES_PASS}"
AUTH_DB="${AUTH_POSTGRES_DATABASE}"

if [ -z "${AUTH_USER}" ] || [ -z "${AUTH_PASSWORD}" ] || [ -z "${AUTH_DB}" ]; then
    echo "Error: Required environment variables are not set."
    exit 1
fi

psql -v ON_ERROR_STOP=1 --username "postgres"  <<-EOSQL
    CREATE USER "$AUTH_USER" WITH PASSWORD '$AUTH_PASSWORD';
EOSQL

psql -v ON_ERROR_STOP=1 --username "postgres"  <<-EOSQL
    CREATE DATABASE "$AUTH_DB";
    GRANT ALL PRIVILEGES ON DATABASE "$AUTH_DB" TO "$AUTH_USER";
EOSQL

# Connect to the newly created database and create the table
psql -v ON_ERROR_STOP=1 --username "postgres" --dbname "$AUTH_DB" <<-EOSQL
    GRANT ALL PRIVILEGES ON SCHEMA public TO ${AUTH_USER};

    CREATE TABLE auth_table (
        id SERIAL PRIMARY KEY,
        username varchar(255) UNIQUE NOT NULL,
        email varchar(255) UNIQUE NOT NULL,
        password varchar(255) NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT NOW()
    );

    GRANT SELECT, INSERT, UPDATE, DELETE ON auth_table TO ${AUTH_USER};
    GRANT USAGE, SELECT ON SEQUENCE auth_table_id_seq TO ${AUTH_USER};
EOSQL

