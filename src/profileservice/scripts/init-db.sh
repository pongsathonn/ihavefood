#!/bin/bash
set -e

PROFILE_USER="${PROFILE_POSTGRES_USER}"
PROFILE_PASSWORD="${PROFILE_POSTGRES_PASS}"
PROFILE_DB="${PROFILE_POSTGRES_DATABASE}"

psql -v ON_ERROR_STOP=1 --username "postgres"  <<-EOSQL
    CREATE USER "$PROFILE_USER" WITH PASSWORD '$PROFILE_PASSWORD';
EOSQL

psql -v ON_ERROR_STOP=1 --username "postgres"  <<-EOSQL
    CREATE DATABASE "$PROFILE_DB";
    GRANT ALL PRIVILEGES ON DATABASE "$PROFILE_DB" TO "$PROFILE_USER";
EOSQL

psql -v ON_ERROR_STOP=1 --username "postgres" --dbname "$PROFILE_DB" <<-EOSQL
    GRANT ALL PRIVILEGES ON SCHEMA public TO "$PROFILE_USER";

    CREATE TABLE profiles (
        user_id VARCHAR(255),
        username VARCHAR(255) UNIQUE NOT NULL,     
        bio TEXT,                                  
        facebook VARCHAR(255),                     
        instagram VARCHAR(255),                    
        line VARCHAR(255),                         
        create_time TIMESTAMP NOT NULL DEFAULT NOW(),
        update_time TIMESTAMP NOT NULL DEFAULT NOW(),
        PRIMARY KEY (user_id)
    );

    CREATE TABLE addresses (
        address_id INT GENERATED ALWAYS AS IDENTITY,
        user_id VARCHAR(255),
        address_name VARCHAR(255),                 
        sub_district VARCHAR(255),                 
        district VARCHAR(255),
        province VARCHAR(255),
        postal_code VARCHAR(20),
        PRIMARY KEY(address_id),
        CONSTRAINT fk_profile
            FOREIGN KEY(user_id)
            REFERENCES profiles(user_id)
            ON DELETE CASCADE
    );

    GRANT SELECT, INSERT, UPDATE, DELETE ON profiles TO "$PROFILE_USER";
    GRANT SELECT, INSERT, UPDATE, DELETE ON addresses TO "$PROFILE_USER";
EOSQL




