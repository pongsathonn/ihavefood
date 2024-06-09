TODO use dirname pubsub

Algorithm are used in this project
- Dijkstra algorithm
- Sorting algorithm
- Dynamic programming Algorithm


Tech stacks
- Redis : caching
- RabbitMQ: Messsage queue, AMQP



" -------------------------------

I think this concept it quit bad, There are alot of Logic that wired with other service
Shall we reponsible logic only its service ?
I think each service just service endpoint .

- User : user, users
- Food : Food, Foods
- Coupon : Coupons, Selected Coupon
- Auth : User Authen
- Payment : Transaction

CQRS design pattern
Read and Write seperate 
Read = Doing Querying such as Query User, Query ListFood. Generally Read from NoSQL (Mongo)
Wirte = Doing with Command such as UpdateFoodCommand . May be with Relational DB ( Postgres ) or NoSQL

Event = Occured after Command done Such as UpdatedFoodCommand then Create FoodUpdatedEvent

So to make our Database Read and Write Sync. We need to use Event driven architecture
When UpdatedEvent is created ( from command ) Our Read Database ( Mongo) Subscribe this Event and Sync together


And Logic should handle in API composition ( API Gateway) ?
Tomorrow read more microservice blog

" --------------------------


title The Best Food-delivery App that have exists

actor Client
participant API Gateway
participant Event Bus

participant Order
participant Payment
participant Delivery
participant Restaurant
participant Coupon
participant User

database DB

lifelinestyle #grey:0.4

==TODO NotificationService Long run connection==

==Process Order Food==
activate API Gateway

// Client send PlaceOrderRequest
Client->>API Gateway:REST : PlaceOrderCommand \nPayload { Foods, Coupon, Address, Contact }

//validate restaurant
API Gateway->>Restaurant: CheckMenuAvaliable
activate Restaurant
Restaurant->API Gateway: response
deactivate Restaurant

//validate Coupon
API Gateway->>Coupon: CheckCouponValid
activate Coupon
Coupon->API Gateway: response
deactivate Coupon

// Order Service
API Gateway->Order: PlaceOrderCommand
activate Order
Order->DB:save place order
activate DB
DB->Order: ok
deactivate DB
Order->API Gateway:Response order status\nPENDING \nPlaceOrderResponse : order_tracking_id
API Gateway->Client: Notify Order Status\nPENDING

// Order: OrderPlacedEvent
Order-#red>Event Bus:OrderPlacedEvent
deactivate Order
activate Event Bus #lightgrey

// Payment
Event Bus-#red>>Payment:OrderPlacedEvent\n body {PAYMENT_METHOD} 
activate Payment
Payment->Payment: if promtpay send to User\nif Credit card do something
Payment->DB:User : Paid\nSave Order payment transaction
activate DB
DB->Payment:saved transaction
deactivate DB
Payment-#blue>Event Bus:OrderPaidEvent
Payment->API Gateway: payment status:PAID
deactivate Payment
API Gateway->Client:Notify Payment Status\nPAID

// Order : update PaymentStatus
Event Bus-#blue>Order:OrderPaidEvent
activate Order
Order->DB:updateOne: PaymentStatus:Paid
activate DB
DB->Order:updated
deactivate DB
deactivate Order

// Order is confirmed by Restaurant
Client-->API Gateway:Restaurant Received Order\nbody :order_id
API Gateway->Order: OrderReceivedCommand
activate Order
Order->Order:Update Orderstatus\nPREPARING_ORDER
Order->API Gateway: Respose Push notification
deactivate Order
API Gateway->Client:Notify Order Status \nPREPARING_ORDER


// Delivery : find rider 
Event Bus-#red>>Delivery: OrderPlacedEvent 
activate Delivery
Delivery->Delivery:Find Rider
Delivery->API Gateway:Send Push Notification
API Gateway->Client:Notify Order Status\nFINDING_RIDER
Delivery-#green>Event Bus: OrderAssignedEvent



API Gateway->Delivery:send 

Delivery->API Gateway:Found Rider\nOrder Status : WAITING_YOUR_RESTAURANT
deactivate Delivery
API Gateway->Client:Notify Order Status \nWAITING_YOUR_RESTAURANT


Event Bus-#green>Order: OrderAssignedEvent
activate Order
Order->DB:updateOne status: ONGOING
activate DB
DB->Order:ok
deactivate DB

--------------------------------------------------------------

How to run genproto.sh ?  
# assume that you have install go grpc plugins if not run this command 
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

1) give permission to script file $ chmod +x genproto.sh  
2) execute $ ./genproto.sh


---------------------

# cURL test

curl =X POST \
=d '{
  "username": "ronaldo",
  "order_cost": 150,
  "menus": [
    {"food_name": "Pad Thai", "price": "50","avaliable":true},
    {"food_name": "Tom Yum Goong", "price": "70","avaliable":true}
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
  "payment_method":0
}' http://localhost:12360/api/orders/place=order



#---------------------------------------------------------------------

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


+================================================================================+
|   exchange    | type   |   routing key                  |    queue             |
|================================================================================|
|   order       | topic  |   Order.Placed.Event           |    payment queue     |  create order transction
|   order       | topic  |   Order.Placed.Event           |    coupon queue      |  update coupon
|   order       | topic  |   Order.Placed.Event           |    delivery queue    |  finding rider
|   order       | topic  |   Order.Confirmed.Event        |                      |  Order success
|               |        |                                |                      |
|   order       | topic  |   Order.Cancelled.Event        |    coupon queue      |   cancel coupon used  to avaliable
|   order       | topic  |   Order.Cancelled.Event        |    delivery queue    |   Cancel delivery order
|               |        |                                |                      |
|               |        |                                |                      |
|   Delivery    | topic  |   OnGoing.Order.Event          |    order queue       |   update order status ( ONGOING )
|   Delivery    | topic  |   WaitingRestau.Order.Event    |    order queue       |   update order status ( WAITING_YOUR_RESTAURNT)
|   Delivery    | topic  |   Order.Assigned.Event         |    order queue       |   
|               |        |                                |                      |
|               |        |                                |                      |
|   Payment     | topic  |   Payment.Paid.Event           |    order queue       |   update order_payment status
|   Payment     | topic  |   Payment.Failed.Event         |    order queue       |   
|               |        |                                |                      |
|               |        |                                |                      |
|   User        | topic  |   User.Updated.Command         |    order queue       |   update username or email
|   User        | topic  |   User.Deleted.Command         |    order queue       |   delete username or email
|               |        |                                |                      |
|   Restaurant  | topic  |   Restaurant.                  |                      |
|               |        |                                |                      |
|               |        |                                |                      |
|               |        |                                |                      |
|               |        |                                |                      |
+================================================================================+


Routing Key 

Order Event
   < service > . < past verb > . < event >  
   ie. Order.Placed.Event , Restaurant.Received.Event, Order.Assigened.Event, Order.Confirmed.Event 
   Order.Cancelled.Event

Order Status
   < status > . < order > . < event >  
   ie. OnGoing.Order.Event , Paid.Order.Event, Confirmed.Order.Event 


User order food pseudo code

1. Client send "PlaceOrderCommand" to OrderService
2. OrderService response "PlaceOrderResponse" alongwith create "OrderPlacedEvent
3. CouponService subscribe to "OrderPlaceEvent" and update coupon in database 
4. DeliveryService subscribe to "OrderPlaceEvent" and then finding rider then publish "OnGoingOrderEvent" 
5. PaymentService subscribe to "OrderPlaceEvent" and create order transaction  then publish "PaidOrderEvent"

OrderService will subscribe to " *.Order.Event " i.e "OnGoing.Order.Event, "Paid.Order.Event"

6. OrderService subscribe to "Paid.Order.Event" then update payment_status
7. Rider or User received order then Client send OrderConfirmCommand to OrderService
8. OrderService update order status then publish "OrderConfirmedEvent"




