
# CURL TEST
this file contains every endpoint test with curl every services

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

# Restaurant 
<pre>

<b>List Menus</b>
curl -X GET http://localhost:180/api/restaurants/{restaurant_name} \
-H "Authorization: Bearer <your_token_here>"

<b>List Restaurants</b>
curl -X GET http://localhost:180/api/restaurants \
-H "Authorization: Bearer <your_token_here>"

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

<b> Get User Profile </b>
curl -X GET http://localhost:180/api/users/{username} \
-H "Authorization: Bearer <your_token_here>"

<b> List User Profile </b>
curl -X GET http://localhost:180/api/users \
-H "Authorization: Bearer <your_token_here>"

<b> Delete User Profile </b>
curl -X GET http://localhost:180/api/users/{username} \
-H "Authorization: Bearer <your_token_here>"

</pre>

<!------------------------------------------------------------------------>

# Auth Service
<pre>

<b>Register</b>
curl -X POST "http://localhost:180/auth/register" \
-H "Content-Type: application/json" \
-d "{
    "username": "___",
    "email": "____",
    "password": "____",
    "phone_number": "___",
    "address": {
        "address_name": "123 Sukhumvit Road",
        "sub_district": "Khlong Toei",
        "district": "Khlong Toei",
        "province": "Bangkok",
        "postal_code": "10110"
    }
}"


<b>Login</b>
curl -s -X POST "http://localhost:180/auth/login" \
-H "Content-Type: application/json" \
-d "{"username":"________", "password":"_______"}")

</pre>

<!------------------------------------------------------------------------>

# Coupon Service
<pre>

<b>Get coupon</b>
curl -s -X GET "http://localhost:180/api/coupons/{code}" \
-H "Authorization: Bearer <your_token_here>"

<b>List coupon</b>
curl -s -X GET "http://localhost:180/api/coupons" \
-H "Authorization: Bearer <your_token_here>"



</pre>


<!------------------------------------------------------------------------>



