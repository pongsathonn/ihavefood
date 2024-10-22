#!/bin/bash
set -e

# this script will be executed in container

USER_USER="${USER_POSTGRES_USER}"
USER_PASSWORD="${USER_POSTGRES_PASS}"
USER_DB="${USER_POSTGRES_DATABASE}"

if [ -z "${USER_USER}" ] || [ -z "${USER_PASSWORD}" ] || [ -z "${USER_DB}" ]; then
    echo "Error: Required environment variables are not set."
    exit 1
fi


psql -v ON_ERROR_STOP=1 --username "postgres"  <<-EOSQL
    CREATE USER "$USER_USER" WITH PASSWORD '$USER_PASSWORD';
EOSQL

psql -v ON_ERROR_STOP=1 --username "postgres"  <<-EOSQL
    CREATE DATABASE "$USER_DB";
    GRANT ALL PRIVILEGES ON DATABASE "$USER_DB" TO "$USER_USER";
EOSQL

psql -v ON_ERROR_STOP=1 --username "postgres" --dbname "$USER_DB" <<-EOSQL
    GRANT ALL PRIVILEGES ON SCHEMA public TO "$USER_USER";

    CREATE TABLE profile (
        id SERIAL PRIMARY KEY,
        username VARCHAR(255) UNIQUE NOT NULL,     
        picture BYTEA,                             
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

    GRANT SELECT, INSERT, UPDATE, DELETE ON profile TO "$USER_USER";
    GRANT USAGE, SELECT ON SEQUENCE profile_id_seq TO ${USER_USER};
EOSQL




