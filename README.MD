# IHAVEFOOD

ihavefood is a microservice food-delivery project written in Go .

![xd](design.png)


Tech Stack and Tools
- Languages: Go, SQL, Shell Script
- Communication Protocols: REST, gRPC, gRPC-Gateway
- Databases: PostgreSQL, MongoDB
- Message Brokers: RabbitMQ
- Caching : TODO 
- Containerization: Docker
- Authentication: JWT, mTLS(TODO)
- Development Environments : Linux(Ubuntu)
- API Documentation : Swagger/OpenAPI



RabbitMQ conventions for every services
- exchange name: "my_exchange"
- exchange type: "direct"

 
### Event Routing Table (Order processing)

<pre>

 | Publisher            |     Routing Key           |           Queue              | Subscriber |  ORDER STATUS  |
 |----------------------|---------------------------|------------------------------|------------|----------------|
 |                      |                           |                              |            | PENDING        |
 | Order                | order.placed.event        | restaurant_assign_queue      | Restaurant |    "           | 
 | Order                | order.placed.event        | rider_assign_queue           | Delivery   |    "           | 
 |                      |                           |                              |            |                |
 | Delivery             | rider.finding.event       | order_status_update_queue    | Order      | FINDING_RIDER  | 
 |                      |                           |                              |            |                |
 | Restaurant           | restaurant.accepted.event | order_status_update_queue    | Order      | PREPARING      | 
 |                      |                           |                              |            |                |
 | Restaurant           | food.ready.event          | order_status_update_queue    | Order      | WAIT_PICKUP    | 
 |                      |                           |                              |            |                |
 | Delivery             | rider.assigned.event      | order_status_update_queue    | Order      | ONGOING        | 
 |                      |                           |                              |            |                |
 | Payment or Delivery  | order.paid.event          | payment_status_update_queue  | Order      |    "           | 
 | Payment or Delivery  | order.paid.event          | coupon_update_queue          | Coupon     |    "           |
 |                      |                           |                              |            |                |
 | Delivery             | rider.delivered.event     | order_status_update_queue    | Order      | DELIVERED      |

</pre>


### Canceling Order 

 | Publisher  |       Routing Key                | Subscriber |  STATUS   |
 |------------|----------------------------------|------------|-----------|
 | Payment    | error.payment.failed.event       |     *      | CANCELLED | 
 | Delivery   | error.rider.unaccepted.event     |     *      | CANCELLED | 
 | Restaurant | error.restaurant.cancelled.event |     *      | CANCELLED | 



# User
The application manages user with two services AuthService and UserService. AuthService 
handles user credentials, whereas UserService manages the user profile.To keep user information, 
such as UserID and Username, synchronized between both services,I use gRPC communication directly
between AuthService and UserService when user information is manipulated.

## Why Separate User  into AuthService and ProfileService?
Initially, when I built my application, I had a single UserService that managed all user-related 
information.However, a problem arose when I integrated JWT into the application.

When a client needs to use an API to access resources, the token must be validated first. This works 
fine for other services because the gateway can validate the token with UserService before performing 
any further validations. But what happens if the client needs to use UserService's API? How will the 
gateway validate the token in that case? It can't call UserService to validate the token because token 
validation must happen first.

To solve this issue, I decided to split UserService into two separate services:
- AuthService: Handles authentication-related operations and token validation. It is coupled with the 
gateway for authentication purposes.
- ProfileService: Manages user profile information and synchronizes user data with AuthService.


# API design
follow some of Google api design <https://cloud.google.com/apis/design>

## gRPC methods example

| Verb   | Noun | Method Name      | Request Message         | Response Message          |
|--------|------|------------------|-------------------------|---------------------------|
| List   | Book | ListBooks        | ListBooksRequest        | ListBooksResponse         |
| Get    | Book | GetBook          | GetBookRequest          | Book                      |
| Create | Book | CreateBook       | CreateBookRequest       | Book                      |
| Update | Book | UpdateBook       | UpdateBookRequest       | Book                      |
| Rename | Book | RenameBook       | RenameBookRequest       | RenameBookResponse        |
| Delete | Book | DeleteBook       | DeleteBookRequest       | google.protobuf.Empty     |

When a client performs a write operation (such as creating or updating a resource), it often
expects the full resource data returned upon completion of that operation. However, most databases
return only basic information, like the newly created resource ID, instead of the full resource
details. To meet the client's expectation, the application layer must perform an additional read 
operation to retrieve the full resource data, resulting in the following sequence.

This design not only ensures that the client's expectations are met but also provides a cleaner
separation of concerns for database operations "READ" and "WRITE".

Example flow

```
Client -> CreateBookRequest -> Server -> Create{request} -> Database
                               Server <-     ID          <- Database
                               Server -> Find{ID}        -> Database
Client <-      Book         <- Server <-    Book         <- Database

```

"READ" returns rerource.
"WRITE" returns identifier(ID).

### database methods example

| Verb      | Noun | Method Name      | Request Message         | Response |
|-----------|------|------------------|-------------------------|----------|
| List      | User | Users            | FindsUserRequest(filter)| Users    |
| Get       | User | User             | UserID                  | User     |
| Create    | User | Create           | NewUserRequest          | UserID   |
| Update    | User | Update           | UserID,User             | UserID   |
| Update    | User | UpdateUserName   | UserID,newName          | UserID   |
| Delete    | User | DeleteUser       | DeleteUserRequest       |          |
| Check     | User | CheckExists      | UserName | Email        | boolean  |
| Validate  | User | ValidateLogin    | UserName,Password       | boolean  |


NOTE: Method Update updates many fields at once

- Check to ensure no duplicate resources or fields exist.
- Validate inputs, such as login credentials or token authenticity.



# Database


My application use two databvase PostgersDB and MongoDB.

Postgers is used wuth AuthService and UserService,other service use Mongo

In the Application Layer, I use structs generated from protobuf complier(protoc) . 
For the database access layer, add the prefix "db" to distinguish them.

```
 Protobuff message           App Layer                            Data Layer                
                                                                                                 
 message Book {              type Book struct {                   type dbBook struct {          
     string id = 1;              ID    string  `protobuf:"x"`           ID    string   
     string name = 2;            Name  string  `protobuf:"x"`           Name  string   
     Color color = 3;            Color Color   `protobuf:"x"`           Color dbColor  
 }                           }                                    }                            

```


# Time
My application is based in Thailand, and the local timezone is Asia/Bangkok. For reading logs within 
the application, focusing on time differences is crucial, and Unix time is not ideal for monitoring 
due to its lack of readability. To establish a consistent time standard throughout the application, 
I have implemented the following time-handling strategy:

Application and Data Layer:
- Use the "time.Time" type in Go for handling timestamps.
- For generating timestamps, use "time.Now()", as it generates time in the local timezone (Asia/Bangkok).
- When inserting into the database (e.g., PostgreSQL or MongoDB), pass the time.Time value directly. 
These database drivers handle the time.Time type in Go seamlessly.

Transport Layer:
- Use the "timestamppb.Timestamp" type, which is generated from "google.protobuf.Timestamp" for gRPC responses.
- When using gRPC-Gateway, the timestamp will be automatically converted to UTC date-time format for the client.


# Error TODO improve doc

error from initialization use log.Fatal() for connection might try to retry first then use
log.Fatal. 

For error response to client the first thing i need to aware is return sensitive information.
to prevent this case happen, always return error messsage that i know what message will be
such as if i check user information in database and i know error from this is "notfound then
an error must be something like "username not found", but if i don't know what error will be
return then i will use internal error which message contains what i do. such as 





