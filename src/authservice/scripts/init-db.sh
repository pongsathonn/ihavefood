#!/bin/bash
set -e

# fetch env variable from .env file
AUTH_USER="${AUTH_POSTGRES_USER}"
AUTH_PASS="${AUTH_POSTGRES_PASS}"
AUTH_DB="${AUTH_POSTGRES_DATABASE}"

psql -v ON_ERROR_STOP=1 --username "postgres"  <<-EOSQL
    CREATE USER "$AUTH_USER" WITH PASSWORD '$AUTH_PASS';
EOSQL

psql -v ON_ERROR_STOP=1 --username "postgres"  <<-EOSQL
    CREATE DATABASE "$AUTH_DB";
    GRANT ALL PRIVILEGES ON DATABASE "$AUTH_DB" TO "$AUTH_USER";
EOSQL

# Connect to the newly created database and create the table
psql -v ON_ERROR_STOP=1 --username "postgres" --dbname "$AUTH_DB" <<-EOSQL

    CREATE TABLE user_credentials (
        id SERIAL PRIMARY KEY,
        username VARCHAR(255) UNIQUE NOT NULL,
        email VARCHAR(255) UNIQUE NOT NULL,
        password VARCHAR(255) NOT NULL,
        role SMALLINT NOT NULL,
        phone_number UNIQUE VARCHAR(255),
        create_time TIMESTAMP NOT NULL DEFAULT NOW()
    );

    GRANT SELECT, INSERT, UPDATE, DELETE ON user_credentials TO "$AUTH_USER";
    GRANT USAGE, SELECT ON SEQUENCE user_credentials_id_seq TO "$AUTH_USER";
EOSQL

