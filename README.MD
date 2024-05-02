
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


title Food-delivery App

actor Client
participant API Gateway
participant Event Bus

participant Order
participant Payment
participant Delivery
participant Restaurant
participant Coupon
participant Notification
participant Review
participant User

database Cache
database DB

lifelinestyle #grey:0.4


Client->>API Gateway:REST : PlaceOrderCommand \nPayload { Foods, Coupon, Address, Contact }
activate API Gateway
activate Event Bus #lightgrey
API Gateway->Order:REST : PlaceOrderCommand

activate Order
Order->Restaurant:gRPC: ValidatePlaceOrder
activate Restaurant
Restaurant->DB:TODO:Call Cache, Save to DB
Restaurant->DB:ValidateFoodExists
activate DB
DB-->Restaurant:Exists
deactivate DB
Restaurant-->Order:Valid
deactivate Restaurant
Order-->Event Bus:OrderFoodEvent
deactivate Order
Event Bus->>Payment:OrderFoodEvent
activate Payment
Payment->Payment:Send QRcode\nTo Client \nExpired 30 minutes

Payment-->Event Bus:PaymentSuccessEvent
deactivate Payment

Event Bus->>Restaurant:OrderFoodEvent
activate Restaurant
Restaurant->Restaurant:Cooking
Restaurant-->Event Bus:OnCookingEvent
deactivate Restaurant

Event Bus->>Coupon:OrderFoodEvent
activate Coupon
Coupon->DB:UpdateUsedCoupon
activate DB
DB-->Coupon:Successful
deactivate DB
Coupon-->Event Bus:UpdatedCoupon
deactivate Coupon
Event Bus->>Restaurant:OrderFoodEvent
activate Restaurant
Restaurant->DB:UpdateFood
activate DB
DB-->Restaurant:Successful
deactivate DB
Restaurant-->Event Bus:Updated
deactivate Restaurant
Event Bus->>Delivery:OrderFoodEvent
activate Delivery
Delivery->Delivery:NotifyToRider

Delivery-->Event Bus: DeliveryOngoingEvent
Delivery->Delivery:Client Confirmed\nReceived Order
Delivery-->Event Bus:OrderReceivedEvent
deactivate Delivery


Event Bus->>Notification:PaymentSuccessEvent
activate Notification
Notification->Notification:Notify:OrderHasPaid
Event Bus->>Notification:OnCookingEvent
Notification->Notification:Status:OnCooking
Event Bus->>Notification: DeliveryOngoingEvent
Notification->Notification:OngoingDeliver
deactivate Notification

Event Bus->Review:OrderReceivedEvent
activate Review
Review-->API Gateway: REST :OrderReviewForm
deactivate Review
API Gateway-->Client:OrderReviewForm


--------------------------------------------------------------

How to run genproto.sh ?  
# assume that you have install go grpc plugins if not run this command 
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

1) give permission to script file $ chmod +x genproto.sh  
2) execute $ ./genproto.sh



