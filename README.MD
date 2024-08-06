#I HAVE FOOD

IHAVEFOOD is a food-delivery microservice personal app written by Go 

Eventbus use Rabbitmq






<pre>

TODO use dirname pubsub

Algorithm are used in this project
- Dijkstra algorithm
- Sorting algorithm
- Dynamic programming Algorithm

-- Priority Queue (Heap) for managing order processing based on priority.
-- Trie (Prefix Tree) for efficient menu item search and autocomplete.
-- Dijkstra's Algorithm for finding the shortest delivery routes.
-- Hash Table Efficiently store and retrieve key-value pair
-- Dynamic programming (DP) Optimize decisions over time.
-- Binary Search Efficiently find items in sorted lists or arrays

--------------------------------------------------------------

How to run genproto.sh ?  
# assume that you have install go grpc plugins if not run this command 
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

1) give permission to script file $ chmod +x genproto.sh  
2) execute $ ./genproto.sh

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


+==========================================================================+
|  exchange    | type   |   routing key               |    queue           |
|==========================================================================|
|  order       | topic  |   Order.Placed.Event        |    payment queue   |
|  order       | topic  |   Order.Placed.Event        |    coupon queue    |
|  order       | topic  |   Order.Placed.Event        |    delivery queue  |
|  order       | topic  |   Order.Confirmed.Event     |                    |
|              |        |                             |                    |
|  order       | topic  |   Order.Cancelled.Event     |    coupon queue    |
|  order       | topic  |   Order.Cancelled.Event     |    delivery queue  |
|              |        |                             |                    |
|              |        |                             |                    |
|  Delivery    | topic  |   OnGoing.Order.Event       |    order queue     |
|  Delivery    | topic  |   WaitingRestau.Order.Event |    order queue     |
|  Delivery    | topic  |   Order.Assigned.Event      |    order queue     |
|              |        |                             |                    |
|              |        |                             |                    |
|  Payment     | topic  |   Payment.Paid.Event        |    order queue     |
|  Payment     | topic  |   Payment.Failed.Event      |    order queue     |
|              |        |                             |                    |
|              |        |                             |                    |
|  User        | topic  |   User.Updated.Command      |    order queue     |
|  User        | topic  |   User.Deleted.Command      |    order queue     |
|              |        |                             |                    |
|  Restaurant  | topic  |   Restaurant.               |                    |
|              |        |                             |                    |
+==========================================================================+

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


</pre>


