package pubsub

import (
	"context"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ interface {
	Publish(exchangeName string, routingKey string, body []byte) error
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

func (r *rabbitMQ) Publish(exchangeName string, routingKey string, body []byte) error {
	ch, err := r.conn.Channel()
	if err != nil {
		log.Println(err)
	}
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"order", // name
		"topic", // type
		true,    // durable
		false,   // auto-deleted
		false,   // internal
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "failed to publish")

	err = ch.PublishWithContext(context.TODO(),
		exchangeName, // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		})
	failOnError(err, "failed to publish")

	return nil
}

func (r *rabbitMQ) Subscribe() (<-chan amqp.Delivery, error) {

	routingKey := "order.placed.event"

	ch, err := r.conn.Channel()
	if err != nil {
		log.Println(err)
	}
	failOnError(err, "Failed to open a channel")

	// parameters = name, type, durable, auto-deleted, internal, no-wait, arguments
	err = ch.ExchangeDeclare("order", "topic", true, false, false, false, nil)
	failOnError(err, "Failed to declare an exchange")

	// queue name binding with routing key
	// parameters = name, durable, delete when unusedd, exclusive, no-wait, arguments
	q, err := ch.QueueDeclare("", false, false, true, false, nil)
	failOnError(err, "Failed to declare a queue")

	// parameters = queue name, routing key, exchange
	err = ch.QueueBind(q.Name, routingKey, "order", false, nil)
	failOnError(err, "Failed to bind a queue")

	// parameters = queue, consumer, auto-ack, exclusive, no-local, no-wait, args
	deliveries, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	failOnError(err, "Failed to register a consumer")

	return deliveries, nil
}
