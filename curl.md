
# CURL TEST



AvailabilityStatus
0  = available
1  = unavailiable
2  = unknow

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

<b> </b>


</pre>

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


# User Service
<pre>
<b> Get User </b>
curl -X GET http://localhost:180/api/users/{username} \
-H "Authorization: Bearer <your_token_here>"


</pre>


