
#!/bin/bash
set -e

# CREATE USER donkadmin WITH PASSWORD 'donkpassword';

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE auth_database;
    GRANT ALL PRIVILEGES ON DATABASE auth_database TO donkadmin;
EOSQL

# Connect to the newly created database and create the table
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "auth_database" <<-EOSQL
    GRANT ALL PRIVILEGES ON SCHEMA public TO donkadmin;

    CREATE TABLE auth_table (
        id SERIAL PRIMARY KEY,
        username varchar(255) UNIQUE NOT NULL,
        email varchar(255) UNIQUE NOT NULL,
        password varchar(255) NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT NOW()
    );

    GRANT SELECT, INSERT, UPDATE, DELETE ON auth_table TO donkadmin;

EOSQL

