#!/bin/bash

uri="http://localhost:180"
token="not assign yet"
orderId="not assign yet"

# Function to register
curlRegister(){
    curl -X POST ${uri}/auth/register \
    -H "Content-Type: application/json" \
    -d '{"username":"messi","email":"xxx@example.com","password":"awwwwwww"}'
}

# Function to log in and capture the token
curlLogin(){

    # assign response from curl test to variable
    res=$(curl -s -X POST ${uri}/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"messi", "password":"awwwwwww"}')

    # print actual curl response
    echo ${res}

    # Extract the access token from the response and assign it to the global variable
    token=$(echo ${res} | jq -r ".accessToken")
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

