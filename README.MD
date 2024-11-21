# TODO change doc UserService to ProfileService 

# IHAVEFOOD

ihavefood is a microservice food-delivery project written in Go .

![xd](architecture.png)


Tech Stack and Tools
- Languages: Go, SQL, Shell Script
- Communication Protocols: REST, gRPC, gRPC-Gateway
- Databases: PostgreSQL, MongoDB
- Message Brokers: RabbitMQ
- Caching : TODO 
- Containerization: Docker
- Authentication: JWT(JSON Web Tokens)
- Development Environments : Vim, Linux(Ubuntu)
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



## User
The application manages user with two services AuthService and UserService. AuthService 
handles user credentials, whereas UserService manages the user profile.To keep user information, 
such as UserID and Username, synchronized between both services,I use gRPC communication directly
between AuthService and UserService when user information is manipulated.

# Why Separate UserService into AuthService and ProfileService?
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
- ProfileService: Manages user profile information and synchronizes user data.
This separation ensures a clear distinction of responsibilities and avoids circular dependencies between 
the gateway and UserService.


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

Services in my application use Posgres and Mongo.
For AuthService and UserService use Postgres, other service use Mongo

Postgres driver <https://pkg.go.dev/github.com/lib/pq#section-readme>
Mongo driver <https://pkg.go.dev/go.mongodb.org/mongo-driver/mongo>

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

My application is based in Thailand, and the local timezone is 'Asia/Bangkok'. However, 
converting times to this format can be cumbersome. When reading logs within the application, 
I primarily need to focus on time differences, and Unix time (since it's difficult to read) 
isn't ideal for monitoring. Therefore, to define a consistent time standard for the application, 
I have designed the time handling as follows:

- Transport Layer: I use Unix timestamps (int64) for communication, as they are compatible 
with the int64 type and do not require any special formatting or types, such as timestamp in 
Protobuf. This makes it easier to work with external systems.

- Application and Data Layers: These layers use UTC as the standard time format. Since logs 
and data in the database need to be easily readable and consistent, UTC is the best choice. 
It ensures that time differences across various components are clear and provides a consistent 
format for both internal application logic and database storage.

NOTE: handle int64 from protobuf with time.Unix() in Go.

time.Now() returns local time which does not has conflict with UTC when validate.

```
External(Unix) < ---request/response---> Application(UTC) 
```


# Error

error from initialization use log.Fatal() for connection might try to retry first then use
log.Fatal. 

For error response to client the first thing i need to aware is return sensitive information.
to prevent this case happen, always return error messsage that i know what message will be
such as if i check user information in database and i know error from this is "notfound then
an error must be something like "username not found", but if i don't know what error will be
return then i will use internal error which message contains what i do. such as 







































 

       
        



