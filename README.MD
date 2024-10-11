
ihavefood is a food-delivery microservice personal app written by Go 
Golang, gRPC, HTTP, Docker, MongoDB, PostgreSQL, RabbitMQ
application protobuff message need to manual generate

<pre>

How to run genproto.sh ?  
# assume that you have install go grpc plugins if not run this command 
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

1) give permission to script file $ chmod +x genproto.sh  
2) execute $ ./genproto.sh

#---------------------------------------------------------------------

// All exchange type is topic
// OrderService handle status for order
 +=================================================================================================+
 | Publish    | Exchange |       Routing Key         |      Queue       | Subscribe  |   STATUS    |
 |=================================================================================================|
 |            |          |                           |                  |            | PENDING     | when new order is coming (default order status)
 | Order      | order    | order.placed.event        | restaurant queue | Restaurant |    "        | notify restaurant new order
 | Order      | order    | order.placed.event        | delivery queue   | Delivery   |    "        | save new order to deliverydb
 | Restaurant | order    | order.cooking.event       | order queue      | Order      | PREPARE     | after restaurant accepted order
 | Restaurant | order    | order.cooking.event       | delivery queue   | Delivery   |    "        | start delivery assignemnt (notify to rider)
 | Delivery   | delivery | rider.notified.event      | order queue      | Order      | FIND_RIDER  | update order status to "FIND_RIDER" 
 | Restaurant | order    | order.ready.event         | order queue      | Order      | WAIT_PICKUP | update order status to "WAiT_PICKUP"
 | Delivery   | delivery | rider.accepted.event      | order queue      | Order      | ONGOING     | update order status to "ONGOING"
 | Delivery   | delivery | user.paid.event           | order queue      | Order      |    "        | update payment status to "PAID"
 | Delivery   | delivery | user.paid.event           | coupon queue     | Coupon     |    "        | update coupon being used
 | Order      | order    | order.confirmed.event     |                  |            | DELIVERED   | after user confirmed or rider delivered
 +=================================================================================================+



 Canceling Order 
 +================================================================================================+
 | Publish    | Exchange |       Routing Key          |      Queue       | Subscribe  |  STATUS   |
 |====================================================================================|===========|
 | Order      | order    | payment.failed.event       | order queue      |   *        | CANCELLED | order is unpaid within XX minutes other services
 |   "        |   "      |           "                |      "           |   "        |    "      | need to cancel order process
 | Delivery   | delivery | rider.unaccepted.event     | order queue      | Order      | CANCELLED | rider not found in XX minutes
 | Restaurant | order    | restaurant.cancelled.event | order queue      | Order      | CANCELLED | cancel from restaurant i.e menus unavailable
 | User       | order    | user.cancelled.event       | order queue      | Order      | CANCELLED | cancel from user i.e user reconsider
 +================================================================================================+



</pre>


