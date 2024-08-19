#!/bin/bash

uri="http://localhost:180"
token="not assign yet"
orderId="not assign yet"

# Function to register a new user
curlRegister() {
    # Default values
    local default_username="primmie"
    local default_email="primmieno1world@example.com"

    # Initialize variables with default values
    local UUU="$default_username"
    local MMM="$default_email"

    while getopts "u:m:" opt; do
        case $opt in
            u) UUU="$OPTARG" ;;  # If option is -u, store the argument in USER
            m) MMM="$OPTARG" ;;  # If option is -m, store the argument in EMAIL
            *) echo "Invalid option"; return 1 ;;  # Handle invalid options
        esac
    done

    # Variables to use in curl command
    local userz="$UUU"
    local mailz="$MMM"

    # Perform the curl request
    # NOTE: Use double quotes in the body because it allows for variable expansion
    curl -X POST "${uri}/auth/register" \
    -H "Content-Type: application/json" \
    -d "{
        \"username\": \"$userz\",
        \"email\": \"$mailz\",
        \"password\": \"secret\",
        \"phone_number\": \"091230123\",
        \"address\": {
            \"address_name\": \"123 Sukhumvit Road\",
            \"sub_district\": \"Khlong Toei\",
            \"district\": \"Khlong Toei\",
            \"province\": \"Bangkok\",
            \"postal_code\": \"10110\"
        }
    }"
}


curlLogin() {
    while getopts "u:p:" opt; do
        case $opt in
            u) local username="$OPTARG" ;;  # Store the argument in USER
            p) local password="$OPTARG" ;;  # Store the argument in PASSWORD
            *) echo "Invalid option"; return 1 ;;  # Handle invalid options
        esac
    done

    # Perform the curl request and capture the response
    local res
    res=$(curl -s -X POST "${uri}/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$username\", \"password\":\"$password\"}")

    # Print the actual curl response
    echo "$res"

    # Extract the access token from the response
    token=$(echo "$res" | jq -r ".accessToken")

    # Optionally, you can print or use the token here
    echo "Access Token: $token"
}


curlPlaceOrder(){

    # assign response from curl test to variable
    res=$(curl -s -X POST ${uri}/api/orders/place-order \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${token}" \
    -d '{
      "username": "ronaldo",
      "restaurant_name":"HaiDee",
      "menus": [
        {"food_name": "Pad Thai", "price": 50},
        {"food_name": "Tom Yum Goong", "price": 70}
      ],
      "delivery_fee": 50,
      "coupon_code": "129dh012",
      "coupon_discount": 20,
      "total": 150,
      "address": {
        "address_name": "home",
        "sub_district": "Nong Hoi",
        "district": "Mueang Chiang Mai",
        "province": "Chiang Mai",
        "postalCode": "50000",
        "country": "Thailand"
      },
      "contact": {
        "phone_number": "+668 1234 5678",
        "email": "r7do@mail.com"
      },
      "payment_method": 0
    }' )

    echo ${res}
}

