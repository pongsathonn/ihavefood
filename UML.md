

Sequence UML Diagram for app

-------------------------------
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

