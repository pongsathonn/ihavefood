
# CURL TEST



AvailabilityStatus
0  = available
1  = unavailiable
2  = unknow

# User Service
<pre>

<b> Register </br>
curl -X POST http://localhost:<port>/register \

<b> Login </br>
curl -H "Authorization:<token>" http://localhost:<port>/login

</pre>


# Order Service
<pre>

<b> Place Order </br>
curl -X POST http://localhost:12360/api/orders/place-order \
-d '{
  "username": "ronaldo",
  "restaurant_name":"HaiDee",
  "menus": [
    {"food_name": "Pad Thai", "price": 50},
    {"food_name": "Tom Yum Goong", "price": 70}
  ],
  "delivery_fee":50,
  "coupon_code": "129dh012",
  "coupon_discount": 20,
  "total": 150,
  "address": {
    "address_name": "home",
    "address_info": "123 Sukhumvit Rd",
    "province": "Bangkok"
  },
  "contact": {
    "phone_number": "+668 1234 5678",
    "email": "r7do@mail.com"
  },
  "payment_method":0
}'

</pre>

# Restaurant 

<pre>

<b> Register new restaurant </b>

curl -X POST http://localhost:12360/api/restaurants \
-d '{
  "restaurant_name": "HaiDee",
  "menus": [
    {"food_name": "Pad Thai", "price": 50},
    {"food_name": "Tom Yum Soup", "price": 70},
    {"food_name": "Green Curry", "price": 80}
  ]
}'

<b> List Restaurants</b>
curl -X GET http://localhost:12360/api/restaurants 

<b> List Menus</b>
curl -X GET http://localhost:12360/api/restaurants/{restaurant_name}

<b> Add Menu</b>
curl -X POST http://localhost:12360/api/restaurants/menus \
-d '{
  "restaurant_name": "HaiDee",
  "menus": [
    {"food_name": "Pad Thai", "price": 50}
  ]
}'

</pre>


