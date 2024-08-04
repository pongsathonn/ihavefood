package rabbitmq

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ interface {
	Subscribe() (<-chan amqp.Delivery, error)
}

type rabbitMQ struct {
	conn *amqp.Connection
}

func NewRabbitMQ(conn *amqp.Connection) RabbitMQ {
	return &rabbitMQ{conn: conn}
}

func failOnError(e error, s string) error {
	return fmt.Errorf("%s : %v\n", s, e)
}

func (r *rabbitMQ) Subscribe() (<-chan amqp.Delivery, error) {

	// r.con.CHannel()
	ch, err := r.conn.Channel()
	failOnError(err, "failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"x",      // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "failed to declare an exchange")

	q, err := ch.QueueDeclare(
		"q007", // name
		false,  // durable
		false,  // delete when unused
		true,   // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	failOnError(err, "failed to declare a queue")

	err = ch.QueueBind(
		q.Name, // queue name
		"",     // routing key
		"x",    // exchange
		false,
		nil,
	)
	failOnError(err, "failed to bind a queue")

	deliveries, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Println("failed to register a consumer:", err)
	} else {
		log.Println("Consumer successfully registered for queue:", q.Name)
	}

	return deliveries, nil
}
