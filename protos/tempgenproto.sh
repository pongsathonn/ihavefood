#!/bin/bash

# # TODO add deps for each services

SERVICES=('customerservice' 'authservice' 'merchantservice' 'orderservice' 'couponservice' 'deliveryservice')

service_path=../src/
gateway_path=../api-gateway/

for i in "${!SERVICES[@]}"; do
    out="$service_path${SERVICES[$i]}"
    if [[ $i -eq $((${#SERVICES[@]} - 1)) ]]; then
        out="$gateway_path"
    fi

    protos=("common.proto" "events.proto" "${SERVICES[@]/%/.proto}")
    protoc -I "." \
        --go_out="$out" \
        --go-grpc_out="$out" \
        --grpc-gateway_out="$out" \
        --openapiv2_out=../web/public/openapi \
        --openapiv2_opt=openapi_naming_strategy=simple,allow_merge=true,logtostderr=true,disable_default_errors=true \
        "${protos[@]}"
done

# find ~/workspace/ihavefood -type d -name "genproto" -exec git add {}/* \;
# git commit -m "update protobuf" || echo "No changes to commit"


