#!/bin/bash
set -e

# this script will be executed in container

ROOT_USER="${ROOT_POSTGRES_USER}"
ROOT_PASSWORD="${ROOT_POSTGRES_PASSWORD}"
ROOT_DB="${ROOT_POSTGRES_DATABASE}"

echo "root user" $ROOT_USER
echo "root password" $ROOT_PASSWORD
echo "root database" $ROOT_DB

psql -v ON_ERROR_STOP=1 --username "postgres"  <<-EOSQL
    CREATE USER "$ROOT_USER" WITH PASSWORD '$ROOT_PASSWORD';
EOSQL

psql -v ON_ERROR_STOP=1 --username "postgres"  <<-EOSQL
    CREATE DATABASE "$ROOT_DB";
    GRANT ALL PRIVILEGES ON DATABASE "$ROOT_DB" TO "$ROOT_USER";
EOSQL

# Connect to the newly created database and create the table
psql -v ON_ERROR_STOP=1 --username "postgres" --dbname "$ROOT_DB" <<-EOSQL
    GRANT ALL PRIVILEGES ON SCHEMA public TO ${ROOT_USER};

    CREATE TABLE auth_table (
        id SERIAL PRIMARY KEY,
        username varchar(255) UNIQUE NOT NULL,
        email varchar(255) UNIQUE NOT NULL,
        password varchar(255) NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT NOW()
    );

    GRANT SELECT, INSERT, UPDATE, DELETE ON auth_table TO ${ROOT_USER};
EOSQL


