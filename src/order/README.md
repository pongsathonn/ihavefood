



docker run -d -p 4010:9910 \
--name donk --network mynet \
-e MONGO_USER=donkadmin \
-e MONGO_PASSWORD=donkpassword \
-e MONGO_HOST=donkdb \
-e MONGO_PORT=27017 \
-e AMQP_USER=kenmilez \
-e AMQP_PASSWORD=mypasswordrb \
-e AMQP_HOST=broker-two \
-e AMQP_PORT=5672 \
order-image:v1.0









---------------------------
# cURL test

curl -X POST http://localhost:12360/api/orders/place-order \
-d '{
  "username": "ronaldo",
  "order_cost": 150,
  "restaurant_name":"HaiDee",
  "menus": [
    {"food_name": "Pad Thai", "price": "50"},
    {"food_name": "Tom Yum Goong", "price": "70"}
  ],
  "coupon_code": "129dh012",
  "address": {
    "address_name": "home",
    "address_info": "123 Sukhumvit Rd",
    "province": "Bangkok"
  },
  "contact": {
    "phone_number": "+668 1234 5678",
    "email": "r7do@mail.com"
  },
  "payment_method":2
}'


#=========================

Command,Query,Event

Command = State can be change or not , Typically CREATE UPDATE DELETE
Query = State not change , Typically GET
Event = Occur after state has changed

Command Example
POST    http://foodDelivery.com/api/orders/place-order
PUT     http://foodDelivery.com/api/orders/place-order/newmenu?menu=krapao
DELETE  http://foodDelivery.com/api/orders/place-order/krapao

Query Example
GET    http://foodDelivery.com/api/orders/place-order
GET     http://foodDelivery.com/api/orders/place-order/ronaldo


Place Order Event
+====================================================================+
|   exchange   | type   |     routing key    |         queue         |
|====================================================================|
| order        | topic  | order.placed.event | payment-created-queue |
| order        | topic  | order.placed.event | user=updated=queue    |
+====================================================================+


Order Status Event
+=================================================================+
| exchange     | type   | routing key        | queue              |
|=================================================================|
| order        | topic  | user.event.created | user=created=queue |
| order        | topic  | user.event.updated | user=updated=queue |
| order        | topic  | user.event.deleted | user=deleted=queue |
| order        | topic  | user.cmd.create    | user=create=queue  |
+=================================================================+



