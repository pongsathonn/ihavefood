package pubsub

import (
	"context"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageBroker interface {
	Publish(exchangeName string, routingKey string, body []byte) error
	Subscribe() (<-chan amqp.Delivery, error)
}

type messageBroker struct {
	conn *amqp.Connection
}

func NewMessageBroker(conn *amqp.Connection) MessageBroker {
	return &messageBroker{conn: conn}
}

func failOnError(e error, s string) error {
	return fmt.Errorf("%s : %v\n", s, e)
}

func (mb *messageBroker) Publish(exchangeName string, routingKey string, body []byte) error {
	ch, err := mb.conn.Channel()
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

func (mb *messageBroker) Subscribe() (<-chan amqp.Delivery, error) {

	ch, err := mb.conn.Channel()
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
	failOnError(err, "Failed to declare an exchange")

	//queue name binding with routing key
	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	routingKey := "*.placeOrder.event"
	err = ch.QueueBind(
		q.Name,     // queue name
		routingKey, // routing key
		"order",    // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	failOnError(err, "fail to consume")

	return msgs, nil
}
