#!/bin/bash -eu

<<comment
This script generates Protocol Buffers (protobuf) definitions for a specified service.

### Usage Example:

To generate code for the deliveryservice, run the script with the -s flag
followed by the service number:

    ./thisfile.sh -s 4

comment

# Ensure that you've set the path for protoc
# PATH=$PATH:$GOPATH/bin

# Declare separate arrays for service numbers and names
# Ensure they are in the same order and correspond by index
service_numbers_keys=(1 2 3 4 5 6)
service_names_values=(
    "customerservice"
    "authservice"
    "restaurantservice"
    "deliveryservice"
    "orderservice"
    "couponservice"
)

# Print usage if no flag or invalid flag is provided
usage() {
    echo "Usage: $0 -s <service_number>"
    echo "Available services:"
    # Iterate through the service_numbers_keys to display options
    for i in "${!service_numbers_keys[@]}"; do
        echo "${service_numbers_keys[$i]}: ${service_names_values[$i]}"
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

# Find the service name based on the number
service_name=""
found=false
for i in "${!service_numbers_keys[@]}"; do
    if [[ "${service_numbers_keys[$i]}" -eq "$service_number" ]]; then
        service_name="${service_names_values[$i]}"
        found=true
        break
    fi
done

# Check if service number is valid
if ! $found; then
    echo "Invalid or missing service number."
    usage
fi

# Path to .proto files
protos=..

# delete /genproto if exists
# Using a temporary variable for the full path before removal for clarity
proto_gen_dir="../../src/$service_name/genproto"
if [ -d "$proto_gen_dir" ]; then
    echo "Removing existing directory: $proto_gen_dir"
    rm -rf "$proto_gen_dir"
fi


# Output directory for the selected service
outdir=../../src/$service_name

protoc -I "$protos" \
    --go_out="$outdir" \
    --go-grpc_out="$outdir" \
    --grpc-gateway_out="$outdir" \
    "$protos/common.proto" \
    "$protos/customerservice.proto" \
    "$protos/authservice.proto" \
    "$protos/restaurantservice.proto" \
    "$protos/deliveryservice.proto" \
    "$protos/orderservice.proto" \
    "$protos/couponservice.proto"

# gateway
gateway_gen_dir="../../gateway/genproto"
if [ -d "$gateway_gen_dir" ]; then
    echo "Removing existing directory: $gateway_gen_dir"
    rm -rf "$gateway_gen_dir"
fi

outdir_gateway=../../gateway

protoc -I "$protos" \
    --go_out="$outdir_gateway" \
    --go-grpc_out="$outdir_gateway" \
    --grpc-gateway_out="$outdir_gateway" \
    "$protos/common.proto" \
    "$protos/customerservice.proto" \
    "$protos/authservice.proto" \
    "$protos/restaurantservice.proto" \
    "$protos/deliveryservice.proto" \
    "$protos/orderservice.proto" \
    "$protos/couponservice.proto"

echo "Code generation completed for $service_name and gateway"
