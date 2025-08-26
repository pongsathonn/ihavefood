#!/bin/bash

SERVICES=('customer' 'auth' 'merchant' 'delivery' 'order' 'coupon')

# TODO add deps for each services

path=../src/
gateway_path=../api-gateway/

for s in "${SERVICES[@]}"; do
    s="${s}service"
    protoc -I "." \
        --go_out="$path$s" \
        --go-grpc_out="$path$s" \
        --grpc-gateway_out="$path$s" \
        --openapiv2_out=. \
        --openapiv2_opt=openapi_naming_strategy=simple,allow_merge=true,merge_file_name=foo \
        "common.proto" \
        "customerservice.proto" \
        "authservice.proto" \
        "merchantservice.proto" \
        "deliveryservice.proto" \
        "orderservice.proto" \
        "couponservice.proto"
done



