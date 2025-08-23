#!/usr/bin/env fish

set apples auth customer merchant order coupon delivery gateway
set gateway "../../api-gateway/"
for s in $apples
    set path "../../src/"{$s}"service"

    if test $s = gateway
        set path $gateway
    end

    protoc -I ".." \
        --go_out=$path \
        --go-grpc_out=$path \
        --grpc-gateway_out=$path \
        "../common.proto" \
        "../customerservice.proto" \
        "../authservice.proto" \
        "../merchantservice.proto" \
        "../deliveryservice.proto" \
        "../orderservice.proto" \
        "../couponservice.proto"
end

