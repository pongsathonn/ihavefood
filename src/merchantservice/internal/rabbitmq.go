package internal

import (
	"context"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

type EventHandler struct {
	Queue, Key string
	Handler    func(amqp.Delivery) error
}

type RabbitMQ struct {
	conn *amqp.Connection
}

func NewRabbitMQ(conn *amqp.Connection) *RabbitMQ {
	return &RabbitMQ{conn: conn}
}

func (r *RabbitMQ) Start(handlers []*EventHandler) error {
	for _, handler := range handlers {
		deliveries, err := r.Subscribe(context.Background(), handler.Queue, handler.Key)
		if err != nil {
			return fmt.Errorf("failed to subscribe to %s: %w", handler.Key, err)
		}

		go func(h *EventHandler, deliveries <-chan amqp.Delivery) {
			for msg := range deliveries {
				if err := h.Handler(msg); err != nil {
					slog.Error("handler error", "err", err, "routingKey", msg.RoutingKey)
					continue
				}
			}
		}(handler, deliveries)
		slog.Info("handler started", "queue", handler.Queue, "key", handler.Key)
	}

	select {}
}

func (r *RabbitMQ) Publish(ctx context.Context, key string, msg amqp.Publishing) error {
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
		key,           // routing key
		false,         // mandatory
		false,         // immediate
		msg,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *RabbitMQ) Subscribe(ctx context.Context, queue, key string) (<-chan amqp.Delivery, error) {
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
		key,           // routing key
		"my_exchange", // exchange
		false,         // no-wait
		nil,           // arguments
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
