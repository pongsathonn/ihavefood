

(producer) ---> [exchange] --binding-- [queue]--> (customer)


producer = source service

customer = destination service

exchange = for distribute message to queue

queue    = receiving message from exchange

Binding is relationship between exchange and queue. We can rememeber as
This queue is interested in message from this exchange ( binding )
so we can binding queue and exhange with exchange name right,
But if we need to specific queue for this topic only from exchange 
we'll use routing key

Routing key is just extra parameter to avoid the confusion i.e user.created.event
so a queue can be subscribe from this routing key only

or we can simply read as "we binding queue and exchange with routing key"

# Producer

routing key = <service>.<event>.<verb2>

| exchange     | type   | routing key         |
|---------------------------------------------|
| order        | topic  | order.event.created |
| order        | topic  | order.event.updated |
| order        | topic  | order.event.deleted |
| order        | topic  | order.cmd.create    |


# Customer

queue       = <service>.<event>.<verb2>
binding     = producer routing key

| exchange     | type   |       queue          |
|-----------------------|----------------------|
| order        | topic  |  user-created-queue  |
| order        | topic  |  user-updated-queue  |
| order        | topic  |  user-deleted-queue  |
| order        | topic  |  user-create-queue   |
