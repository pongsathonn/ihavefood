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

    CREATE TABLE profile (
        id VARCHAR(255) PRIMARY KEY,
        username VARCHAR(255) UNIQUE NOT NULL,     
        bio TEXT,                                  
        facebook VARCHAR(255),                     
        instagram VARCHAR(255),                    
        line VARCHAR(255),                         
        address_name VARCHAR(255),                 
        sub_district VARCHAR(255),                 
        district VARCHAR(255),                     
        province VARCHAR(255),                     
        postal_code VARCHAR(20),                   

        create_time TIMESTAMP NOT NULL DEFAULT NOW()
    );

    GRANT SELECT, INSERT, UPDATE, DELETE ON profile TO "$PROFILE_USER";
EOSQL




