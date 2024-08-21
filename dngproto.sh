#!/bin/bash

# this script remove genproto directory from every services and gateway
# and generate new genproto

# services
process_genproto() {
  local service="$1"
  echo "Processing service ... $service "

  rm -rf "./src/$service/genproto" \
    && chmod +x "./src/$service/genproto.sh" \
    && (cd "./src/$service" && ./genproto.sh) \
    || echo "Failed to process genproto for $service"
}

services=("orderservice" "restaurantservice" "couponservice" "deliveryservice")

for service in "${services[@]}"; do
  process_genproto "$service"
done

# services
process_genproto() {
  local service="$1"
  echo "Processing service ... $service "

  rm -rf "./src/$service/genproto" \
    && chmod +x "./src/$service/scripts/genproto.sh" \
    && (cd "./src/$service/scripts" && ./genproto.sh) \
    || echo "Failed to process genproto for $service"
}

services=("userservice" "authservice")

for service in "${services[@]}"; do
  process_genproto "$service"
done


# gateway
echo "Processing gateway ..."
rm -rf "./gateway/genproto" \
  && chmod +x "./gateway/genproto.sh" \
  && (cd "./gateway" && ./genproto.sh) \
  || echo "Failed to process genproto for gateway"

echo "Script execution completed."


