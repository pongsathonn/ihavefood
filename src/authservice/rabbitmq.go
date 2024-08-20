package main

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitmqClient interface {
	Publish(exchange, queue, routingKey string, body []byte) error
	Subscribe(exchange, queue, routingKey string) (<-chan amqp.Delivery, error)
}

type rabbitmqClient struct {
	conn *amqp.Connection
}

func NewRabbitMQ(conn *amqp.Connection) RabbitmqClient {
	return &rabbitmqClient{conn: conn}
}

func (r *rabbitmqClient) Publish(exchange, queue, routingKey string, body []byte) error {
	ch, err := r.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %v", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exchange, // name
		"topic",  // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %v", err)
	}

	err = ch.PublishWithContext(
		context.TODO(),
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish: %v", err)
	}

	return nil
}

func (r *rabbitmqClient) Subscribe(exchange, queue, routingKey string) (<-chan amqp.Delivery, error) {
	ch, err := r.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open the channel: %v", err)
	}

	err = ch.ExchangeDeclare(
		exchange, // name
		"topic",  // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare an exchange: %v", err)
	}

	q, err := ch.QueueDeclare(
		queue, // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare a queue: %v", err)
	}

	err = ch.QueueBind(
		q.Name,     // queue name
		routingKey, // routing key
		exchange,   // exchange
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to bind a queue: %v", err)
	}

	deliveries, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register a consumer: %v", err)
	}

	return deliveries, nil
}
