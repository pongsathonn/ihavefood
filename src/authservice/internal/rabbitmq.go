package internal

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type rabbitMQ struct {
	conn *amqp.Connection
}

func NewRabbitMQ(conn *amqp.Connection) *rabbitMQ {
	return &rabbitMQ{conn: conn}
}

func (r *rabbitMQ) publish(ctx context.Context, routingKey string, msg amqp.Publishing) error {
	ch, err := r.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"my_exchange", // name
		"direct",      // type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(
		ctx,
		"my_exchange", // exchange
		routingKey,    // routing key
		true,          // mandatory
		false,         // immediate
		msg,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *rabbitMQ) subscribe(ctx context.Context, queue, routingKey string) (<-chan amqp.Delivery, error) {
	ch, err := r.conn.Channel()
	if err != nil {
		return nil, err
	}

	err = ch.ExchangeDeclare(
		"my_exchange", // name
		"direct",      // type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	err = ch.QueueBind(
		q.Name,        // queue name
		routingKey,    // routing key
		"my_exchange", // exchange
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		return nil, err
	}

	deliveries, err := ch.ConsumeWithContext(
		ctx,
		q.Name,         // queue
		"auth_service", // consumer
		true,           // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return nil, err
	}

	return deliveries, nil
}
