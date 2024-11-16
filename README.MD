
ihavefood is a food-delivery microservice personal project written in Go .

Technologies in this project
Golang, gRPC, HTTP, Docker, MongoDB, PostgreSQL, RabbitMQ


//  RabbitMQ conventions for every services
// - exchange name: "my_exchange"
// - exchange type: "direct"

 
### Food Delivery Process Table

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
The application manages user with two services AuthService and UserService. 
AuthService handles user credentials, whereas UserService manages the user profile.

To keep user information, such as UserID and Username, synchronized between both services, 
I use gRPC communication directly between AuthService and UserService.


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

My application is based in Thailand, and the local timezone should be 'Asia/Bangkok'.
However, converting times to this format can be cumbersome. When reading logs in the 
application, I primarily need to focus on the time differences, and Unix time is difficult
to read for me. Therefore, to define a consistent standard for the application, 
all time-related values, including responses to clients, should be in 'UTC' format.








 

       
        



