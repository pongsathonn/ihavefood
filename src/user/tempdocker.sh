#!/bin/bash

chmod +x "$(pwd)/init-db.sh"

USER_POSTGRES_PASS=donkpassword

docker run --name pgx -p 5432:5432 \
-e POSTGRES_PASSWORD=$USER_POSTGRES_PASS \
-v "$(pwd)"/init-db.sh:/docker-entrypoint-initdb.d/init-db.sh \
-d postgres:latest



