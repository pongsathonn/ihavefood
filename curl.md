
# CURL TEST

AvailabilityStatus <br>
0  = available <br>
1  = unavailiable <br>
2  = unknow 

<!------------------------------------------------------------------------>
# Order Service
<pre>
<b>Place Order</b>

curl -X POST http://localhost:180/api/orders/place-order \
-H "Content-Type: application/json" \
-H "Authorization: Bearer xxxxx" \
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
}'

<b>User Order History</b>
curl -X GET http://localhost:180/api/orders/{username} \
-H "Authorization: Bearer <your_token_here>"

</pre>

<!------------------------------------------------------------------------>

# Delivery Service
<pre>
curl -X POST http://localhost:180/api/deliveries/accept-order \
-H "Authorization: Bearer <your_token_here>" \
-d '{"rider_id":"rider002", "order_id":"66af00af1687c32893d15693"}'


</pre>

<!------------------------------------------------------------------------>


# Auth Service
<pre>

<b>Register</b>
curl -X POST http://localhost:<port>/register \
-d '{ xxxxxx}'

<b>Login</b>
POST
curl -H "Authorization:<token>" http://localhost:180/login \
-d '{"username":"ken", "password":"secret"}'

</pre>

<!------------------------------------------------------------------------>


# Restaurant 
<pre>
<b>Register new restaurant</b>

curl -X POST http://localhost:180/api/restaurants \
-H "Authorization: Bearer <your_token_here>" \
-d '{
  "restaurant_name": "HaiDee",
  "menus": [
    {"food_name": "Pad Thai", "price": 50},
    {"food_name": "Tom Yum Soup", "price": 70},
    {"food_name": "Green Curry", "price": 80}
  ]
}'

<b>List Restaurants</b>
curl -X GET http://localhost:180/api/restaurants \
-H "Authorization: Bearer <your_token_here>"

<b>List Menus</b>
curl -X GET http://localhost:180/api/restaurants/{restaurant_name} \
-H "Authorization: Bearer <your_token_here>"

<b>Add Menu</b>
curl -X POST http://localhost:180/api/restaurants/menus \
-H "Authorization: Bearer <your_token_here>" \
-d '{
  "restaurant_name": "HaiDee",
  "menus": [
    {"food_name": "Pad Thai", "price": 50}
  ]
}'

</pre>

<!------------------------------------------------------------------------>

# User Service
<pre>
<b> Get User </b>
curl -X GET http://localhost:180/api/users/{username} \
-H "Authorization: Bearer <your_token_here>"

<b> Create New User </b>
curl -X POST http://localhost:180/api/users \
-H "Authorization: Bearer <your_token_here>" \
-H "Content-Type: application/json" \
-d '{
    "username": "_____",
    "email": "_____",
    "phone_number": "",
    "address": {
        "address_name": "_____",
        "sub_district": "_____",
        "district": "_____",
        "province": "_____",
        "postal_code": "_____"
    }
}'

<b> List User </b>
curl -X GET http://localhost:180/api/users \
-H "Authorization: Bearer <your_token_here>"


</pre>

<!------------------------------------------------------------------------>


