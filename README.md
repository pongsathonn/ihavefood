# IHAVEFOOD

ihavefood is a microservice food-delivery project written in Rust,Go.
![xd](design.png)


 
### Event Routing Table (Order processing)

 | Publisher    |     Routing Key           |           Queue              | Subscriber |  ORDER STATUS  |   
 |--------------|---------------------------|------------------------------|------------|----------------|
 |              |                           |                              |            |  PENDING       |
 | Order        | order.placed.event        | restaurant_assign_queue      | Restaurant |                | 
 | Order        | order.placed.event        | rider_assign_queue           | Delivery   |                | 
 | Restaurant   | restaurant.accepted.event | restaurant_accept_queue      | Delivery   |                | 
 |              |                           |                              |            |  PREPARING     |
 | Delivery     | rider.finding.event       | order_status_update_queue    | Order      |                | 
 |              |                           |                              |            |  FINDING_RIDER |
 | Delivery     | rider.assigned.event      | order_status_update_queue    | Order      |                |
 |              |                           |                              |            |  WAIT_PICKUP   |
 | Delivery     | rider.picked_up.event     | order_status_update_queue    | Order      |                | 
 |              |                           |                              |            |  ONGOING       |
 | Delivery     | rider.delivered.event     | order_status_update_queue    | Order      |                |
 |              |                           |                              |            |  DELIVERED     |

# payment

 | Publisher            |     Routing Key           |           Queue              | Subscriber |  PAYMENT STATUS  |   
 |----------------------|---------------------------|------------------------------|------------|------------------|
 |                      |                           |                              |            |  WAIT_PAY        |
 | Delivery             | order.paid.event          | order_paid__queue            | Order      |                  | 
 | Payment    TODO      | order.paid.event          | order_paid_queue             | Delivery   |                  | 
 |                      |                           |                              |            |  PAID            |
 | Delivery             | order.paid.event          | coupon_update_queue          | Coupon     |                  |


### Canceling Order 
# proceed to refund

 | Publisher  |       Routing Key                | Subscriber |  STATUS   |
 |------------|----------------------------------|------------|-----------|
 | Payment    | error.payment.failed.event       |     *      | CANCELLED | 
 | Delivery   | error.rider.unaccepted.event     |     *      | CANCELLED | 
 | Restaurant | error.restaurant.cancelled.event |     *      | CANCELLED | 
