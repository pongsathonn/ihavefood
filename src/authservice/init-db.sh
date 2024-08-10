#!/bin/bash
set -e

# this script will be executed in container

USER="${POSTGRES_USER}"
DB="${POSTGRES_DATABASE}"

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE "$DB";
    GRANT ALL PRIVILEGES ON DATABASE "$DB" TO "$USER";
EOSQL

# Connect to the newly created database and create the table
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$DB" <<-EOSQL
    GRANT ALL PRIVILEGES ON SCHEMA public TO ${USER};

    CREATE TABLE auth_table (
        id SERIAL PRIMARY KEY,
        username varchar(255) UNIQUE NOT NULL,
        email varchar(255) UNIQUE NOT NULL,
        password varchar(255) NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT NOW()
    );

    GRANT SELECT, INSERT, UPDATE, DELETE ON auth_table TO ${USER};

EOSQL


