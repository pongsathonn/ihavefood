
ihavefood is a food-delivery microservice personal project written in Go .

Technologies in this project
Golang, gRPC, HTTP, Docker, MongoDB, PostgreSQL, RabbitMQ


// All exchange type is topic
// OrderService handle status for order
//
// use exchange "order_x" for entire app
//
// every message body must include orderNO in body
// every services MUST include orderid in body

<pre>
 | Publisher            |     Routing Key           |    Queue     | Subscriber |  ORDER STATUS  |  below this explain what subscriber does
 |----------------------|---------------------------|--------------|------------|----------------|
 |                      |                           |              |            | PENDING        | default order status
 | Order                | order.placed.event        | restaurant.q | Restaurant |    "           | notify restaurant new order
 |    "                 |        "                  | delivery.q   | Delivery   |    "           | send push noti to riders and generate pickup info
 | Delivery             | rider.finding.event       | order.q      | Order      | FINDING_RIDER  | update order status
 | Restaurant           | restaurant.accepted.event | order.q      | Order      | PREPARING      | update order status
 | Restaurant           | food.ready.event          | order.q      | Order      | WAIT_PICKUP    | update order status
 | Delivery             | rider.assigned.event      | order.q      | Order      | ONGOING        | update order status
 | Payment | Delivery   | user.paid.event           | order.q      | Order      |    "           | update payment status to "PAID"
 |    "         "       |        "                  | coupon.q     | Coupon     |    "           | update coupon being used in database
 | Delivery             | rider.delivered.event     | order.q      | Order      | DELIVERED      | update order status to "delivered"
</pre>

### Routing key explanation
  "order.placed.event" The OrderService has received a new order.

  "rider.finding.event" The DeliveryService has started finding a rider.

  "restaurant.accepted.event" The restaurant has accepted the order and has started cooking

  "food.ready.event" The food has been cooked and is ready for pickup by the rider.
  
  "rider.assigned.event" A rider has been successfully assigned to pick up the order.

  "user.paid.event" The user has completed payment for the order.  
    - PaymentService publish this if promptpay or credit card transaction successfully.
    - DeliveryService publish when rider has receive cash from user after delivery.
                     
  "rider.delivered.event" The rider has confirmed that the order has been delivered.


### Canceling Order 

 | Publisher  |       Routing Key                | Subscriber |  STATUS   |
 |------------|----------------------------------|------------|-----------|
 | Payment    | error.payment.failed.event       |     *      | CANCELLED | 
 | Delivery   | error.rider.unaccepted.event     |     *      | CANCELLED | 
 | Restaurant | error.restaurant.cancelled.event |     *      | CANCELLED | 


> this is blockquotes
> also blockquotes

 > The publisher should attach the error message in the body and use a routing 
 key that starts with "error." The errors may include anything critical that
 leads to the cancellation of an order, such as payment transaction failures,
 the inability to find a rider within a specified time, restaurant cancellations,
 and any other internal server errors. Therefore, every service must subscribe 
 to the "error.#" routing key to handle the rollback or deletion of order information.






 




# API design

**"TRY"** to follow Google api design <https://cloud.google.com/apis/design>


Naming should be simple, intuitive and consistent

## gRPC methods

You **are not** late. (indicative)
**Don't be** late! (imperative)

Methods names
- The verb portion of the method name should use the imperative mood, which is for
  orders or commands rather than the indicative mood which is for questions.
- the noun portion of the method name must be singular for all methods except List.
- custome methods may be singular or plural as appropriate.
- Batch methods **must** use the plural noun.

| Verb   | Noun | Method Name      | Request Message         | Response Message          |
|--------|------|------------------|-------------------------|---------------------------|
| List   | Book | ListBooks        | ListBooksRequest        | ListBooksResponse         |
| Get    | Book | GetBook          | GetBookRequest          | Book                      |
| Create | Book | CreateBook       | CreateBookRequest       | Book                      |
| Update | Book | UpdateBook       | UpdateBookRequest       | Book                      |
| Rename | Book | RenameBook       | RenameBookRequest       | RenameBookResponse        |
| Delete | Book | DeleteBook       | DeleteBookRequest       | google.protobuf.Empty     |





### Install protobuff complier

<pre>
 $ go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28 
 $ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
</pre>

## How to execute genproto.sh ?  
1) Give permission to script file $ chmod +x genproto.sh  
2) Execute $ ./genproto.sh 
