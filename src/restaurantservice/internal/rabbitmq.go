package internal

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ interface {
	Publish(ctx context.Context, exchange, routingKey string, body []byte) error
	Subscribe(ctx context.Context, exchange, queue, routingKey string) (<-chan amqp.Delivery, error)
}

type rabbitMQ struct {
	conn *amqp.Connection
}

func NewRabbitMQ(conn *amqp.Connection) RabbitMQ {
	return &rabbitMQ{conn: conn}
}

func (r *rabbitMQ) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	ch, err := r.conn.Channel()
	if err != nil {
		return err
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
		return err
	}

	err = ch.PublishWithContext(
		ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *rabbitMQ) Subscribe(ctx context.Context,
	exchange,
	queue,
	routingkey string,
) (<-chan amqp.Delivery, error) {

	ch, err := r.conn.Channel()
	if err != nil {
		return nil, err
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
		q.Name,     // queue name
		routingkey, // routing key
		exchange,   // exchange
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		return nil, err
	}

	deliveries, err := ch.ConsumeWithContext(
		ctx,
		q.Name,               // queue
		"restaurant_service", // consumer
		true,                 // auto-ack
		false,                // exclusive
		false,                // no-local
		false,                // no-wait
		nil,                  // arguments
	)
	if err != nil {
		return nil, err
	}

	return deliveries, nil
}
