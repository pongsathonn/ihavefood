#!/bin/bash -eu

<<comment
This script generates Protocol Buffers (protobuf) definitions for a specified service.

### Usage Example:

To generate code for the deliveryservice, run the script with the -s flag
followed by the service number:

    ./thisfile.sh -s 4

comment

# Ensure that you've set the path for protoc
PATH=$PATH:$GOPATH/bin

# Declare an associative array mapping service numbers to service names
declare -A services
services=( 
    [1]="profileservice" 
    [2]="authservice" 
    [3]="restaurantservice" 
    [4]="deliveryservice" 
    [5]="orderservice" 
    [6]="couponservice"
)

# Declare an array for ordered service numbers
service_numbers=(1 2 3 4 5 6)

# Print usage if no flag or invalid flag is provided
usage() {
    echo "Usage: $0 -s <service_number>"
    echo "Available services:"
    for num in "${service_numbers[@]}"; do
        echo "$num: ${services[$num]}"
    done
    exit 1
}

# Parse the service number from the command-line argument
while getopts "s:" opt; do
    case ${opt} in
        s)
            service_number=$OPTARG
            ;;
        *)
            usage
            ;;
    esac
done

# Check if service number is valid
if [[ -z "${service_number+x}" || -z "${services[$service_number]:-}" ]]; then
    echo "Invalid or missing service number."
    usage
fi


# Path to .proto files
protos=..

service_name=${services[$service_number]}

# delete /genproto if exists
rm -rf ../../src/$service_name/genproto

# Output directory for the selected service
outdir=../../src/$service_name

protoc -I $protos \
    --go_out=$outdir \
    --go-grpc_out=$outdir \
    --grpc-gateway_out=$outdir \
    $protos/profileservice.proto \
    $protos/authservice.proto \
    $protos/restaurantservice.proto \
    $protos/deliveryservice.proto \
    $protos/orderservice.proto \
    $protos/couponservice.proto

# gateway
rm -rf ../../gateway/genproto
outdir_gateway=../../gateway

protoc -I $protos \
    --go_out=$outdir_gateway \
    --go-grpc_out=$outdir_gateway \
    --grpc-gateway_out=$outdir_gateway \
    $protos/profileservice.proto \
    $protos/authservice.proto \
    $protos/restaurantservice.proto \
    $protos/deliveryservice.proto \
    $protos/orderservice.proto \
    $protos/couponservice.proto

echo "Code generation completed for $service_name and gateway"

