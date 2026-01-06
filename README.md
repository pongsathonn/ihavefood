# IHAVEFOOD

ihavefood is a microservice food-delivery project written in Rust,Go.
![xd](arch.png)



# Order Happy Path

 │ Publisher    │     Routing Key           │           Queue              │ Subscriber │  ORDER STATUS  │   
 ├──────────────├───────────────────────────├──────────────────────────────├────────────├────────────────┤
 │              │                           │                              │            │  PENDING       │
 │ Order        │ order.placed.event        │ merchant_assign_queue        │ Merchant   │                │ 
 │ Order        │ order.placed.event        │ rider_assign_queue           │ Delivery   │                │ 
 │ Merchant     │ merchant.accepted.event   │ merchant_accept_queue        │ Delivery   │                │ 
 │              │                           │                              │            │  PREPARING     │
 │ Delivery     │ rider.notified.event      │ order_status_update_queue    │ Order      │                │ 
 │              │                           │                              │            │  FINDING_RIDER │
 │ Delivery     │ rider.assigned.event      │ order_status_update_queue    │ Order      │                │
 │              │                           │                              │            │  WAIT_PICKUP   │
 │ Delivery     │ rider.picked_up.event     │ order_status_update_queue    │ Order      │                │ 
 │              │                           │                              │            │  ONGOING       │
 │ Delivery     │ rider.delivered.event     │ order_status_update_queue    │ Order      │                │
 │              │                           │                              │            │  DELIVERED     │
 


# Payment

 │ Publisher            │     Routing Key           │           Queue              │ Subscriber │  PAYMENT STATUS  │   
 ├──────────────────────├───────────────────────────├──────────────────────────────├────────────├──────────────────┤
 │                      │                           │                              │            │  WAIT_PAY        │
 │ Delivery             │ order.paid.event          │ order_paid__queue            │ Order      │                  │ 
 │ Payment              │ order.paid.event          │ order_paid_queue             │ Delivery   │                  │ 
 │                      │                           │                              │            │  PAID            │
 │ Delivery             │ order.paid.event          │ coupon_update_queue          │ Coupon     │                  │
 
 
# NOTE: Not impl yet.
 │ Publisher  │       Routing Key                │ Subscriber │  STATUS   │ 
 ├────────────├──────────────────────────────────├────────────├───────────┤
 │ Payment    │ error.payment.failed.event       │     *      │ CANCELLED │  
 │ Delivery   │ error.rider.unaccepted.event     │     *      │ CANCELLED │  
 │ Merchant   │ error.merchant.cancelled.event   │     *      │ CANCELLED │  
 
 

## Data Replication
(Auth act as a source of truth)
Keeping the same data synchronized across multiple services
key = sync.<data>.<verb> e.g, sync.customer.email.updated, sync.rider.phone.deleted

 │ Publisher  │       Routing Key                      │ Subscriber      │ 
 ├────────────├────────────────────────────────────────├─────────────────├
 │ Auth       │ sync.customer.created                  │     Customer    │ 
 │ Auth       │ sync.rider.created                     │     Delivery    │ 
 │ Auth       │ sync.merchant.created                  │     Merchant    │ 
 │ Auth       │ sync.customer.email.updated            │     Customer    │ 
 │ Auth       │ sync.customer.phone_number.updated     │     Customer    │ 
 │ Auth       │ sync.rider.email.updated               │     Delivery    │ 
 │ Auth       │ sync.rider.phone_number.updated        │     Delivery    │ 
 │ Auth       │ sync.merchant.email.updated            │     Merchant    │ 
 │ Auth       │ sync.merchant.phone_number.updated     │     Merchant    │ 

