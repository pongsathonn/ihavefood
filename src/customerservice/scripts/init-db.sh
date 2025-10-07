#!/bin/bash
set -e

CUSTOMER_USER="${CUSTOMER_DB_USER}"
CUSTOMER_PASSWORD="${CUSTOMER_DB_PASS}"
CUSTOMER_DB="${CUSTOMER_DB_NAME}"

psql -v ON_ERROR_STOP=1 --username "postgres"  <<-EOSQL
    CREATE USER "$CUSTOMER_USER" WITH PASSWORD '$CUSTOMER_PASSWORD';
EOSQL

psql -v ON_ERROR_STOP=1 --username "postgres"  <<-EOSQL
    CREATE DATABASE "$CUSTOMER_DB";
    GRANT ALL PRIVILEGES ON DATABASE "$CUSTOMER_DB" TO "$CUSTOMER_USER";
EOSQL

psql -v ON_ERROR_STOP=1 --username "postgres" --dbname "$CUSTOMER_DB" <<-EOSQL
    GRANT ALL PRIVILEGES ON SCHEMA public TO "$CUSTOMER_USER";

    CREATE TABLE customers (
        customer_id UUID,
        username VARCHAR(255) UNIQUE NOT NULL,     
        email VARCHAR(255) UNIQUE NOT NULL,
        facebook VARCHAR(255),                     
        instagram VARCHAR(255),                    
        line VARCHAR(255),                         
        create_time TIMESTAMP NOT NULL DEFAULT NOW(),
        update_time TIMESTAMP NOT NULL DEFAULT NOW(),
        PRIMARY KEY (customer_id)
    );

    CREATE TABLE addresses (
        address_id UUID DEFAULT gen_random_uuid(),
        customer_id UUID,
        address_name VARCHAR(255),                 
        sub_district VARCHAR(255),                 
        district VARCHAR(255),
        province VARCHAR(255),
        postal_code VARCHAR(20),
        PRIMARY KEY(address_id),
        CONSTRAINT fk_profile
            FOREIGN KEY(customer_id)
            REFERENCES customers(customer_id)
            ON DELETE CASCADE
    );

    GRANT SELECT, INSERT, UPDATE, DELETE ON customers TO "$CUSTOMER_USER";
    GRANT SELECT, INSERT, UPDATE, DELETE ON addresses TO "$CUSTOMER_USER";
EOSQL




