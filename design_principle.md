### grpc or amqp ?
use grpc if fetching resources , check validify
use amqp for event,notification

### error and logging message
error return from server shoudl be context boundary otherwise just return internal server error
logging message should be context rather than operation 



