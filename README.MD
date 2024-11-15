
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

When design follow this table. Client expected resources after write methods back.
but almost every database it doesn't return that resoruce back , so application layer
need to call method read again to retrives those resources. 

Example flow

```
[Client]  ->   CreateBookRequest   ->   [ Server  ->     Create{request}   ->   DB Layer ]
[      ]                                [                response{ID}      <-   DB Layer ]
[      ]                                [                Find{ID}          ->   DB Layer ]
[Client]  <-       Book            <-   [ Server  <-     response{Book}    <-   DB Layer ]

```


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





 

       
        



