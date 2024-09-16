
ihavefood is a food-delivery microservice personal app written by Go 

Golang, gRPC, HTTP, Docker, MongoDB, PostgreSQL, RabbitMQ




application protobuff message need to manual generate

<pre>

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

                                           gRPC communication 
 +===========================================================================================================================+  
 |  source service   |          function          |                   description                    |  destination service  | 
 |===========================================================================================================================|
 |     order         |                            |   validate coupon commadn , get coupon discount  |        coupon         |
 |     order         |                            |   get delivery fee                               |        delivery       |
 |                   |                            |                                                  |                       |
 |                   |                            |                                                  |                       |
 |                   |                            |                                                  |                       |
 |                   |                            |                                                  |                       |
 +---------------------------------------------------------------------------------------------------------------------------+

// All exchange type is topic
// RabbitMQ suit for event and command 


// TODO this order.validate.evnet might considerd be command not event
// TODO might delete thir validate envet use grpc instead faster
 +=============================================================================================+
 | Publish | Exchange |       Routing Key         |     Queue        | Subscribe  | STATUS     |        description 
 |================================================================================|=============
 | Order   | order    | order.placed.event        | restaurant queue | Restaurant | PENDING    |  validate restaurant status ( open, closed )
 |         | order    | order.placed.event        | coupon queue     | Coupon     |    "       |  update coupon being used 
 |         | order    | order.placed.event        | delivery queue   | Delivery   |    "       |  start delivery assignment process such as notify to rider
 |         |          |                           |                  |            | PREPARE    |
 |         |          |                           |                  |            |    "       |
 |         |          |                           |                  |            | FIND_RIDER |
 |         |          |                           |                  |            |    "       |
 |         | delivery | order.ongoing.event       | order queue      |            | ONGOING    |
 |         |          |                           |                  |            |    "       |
 |         |          |                           |                  |            |            | <= Above this valid
 |         | delivery | OnGoing.Order.Event       | order queue      |            |            |
 |         | delivery | WaitingRestau.Order.Event | order queue      |            |            |
 |         | delivery | Order.Assigned.Event      | order queue      |            |            |
 |         |          |                           |                  |            |            |
 |         |          |                           |                  |            |            |
 |         | payment  | Payment.Paid.Event        | order queue      |    All     |            |  To confirm order success , every services can committ
 |         |          |                           |                  |            |            |
 |         |          |                           |                  |            |            |
 | <----------------------------------( AFTER USER CONFIRMED)---------------------- DELIVERED->| 
 |         | user     | User.Updated.Command      | order queue      |            |            |
 |         | user     | User.Deleted.Command      | order queue      |            |            |
 |         |          |                           |                  |            |            |
 |         |          |                           |                  |            |            |
 |         | payment  | Payment.Failed.Event      | order queue      |            |            |
 |         | order    | Order.Cancelled.Event     | coupon queue     |            |            |
 |         | order    | Order.Cancelled.Event     | delivery queue   |            |            |
 |         |          |                           |                  |            |            |
 +=============================================================================================+


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


