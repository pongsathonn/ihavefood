#!/bin/bash
set -e

# CREATE USER donkadmin WITH PASSWORD 'donkpassword';

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE user_database;
    GRANT ALL PRIVILEGES ON DATABASE user_database TO donkadmin;
EOSQL

# Connect to the newly created database and create the table
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "user_database" <<-EOSQL
    GRANT ALL PRIVILEGES ON SCHEMA public TO donkadmin;

    CREATE TABLE user_table (
        id SERIAL PRIMARY KEY,
        username varchar(255) UNIQUE, 
        email varchar(255),
        password varchar(255),
        phone_number varchar(20),
        address_name varchar(255),
        address_info varchar(255),
        province varchar(255)
    );

    GRANT SELECT, INSERT, UPDATE, DELETE ON user_table TO donkadmin;

EOSQL

