package internal

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ interface {
	Publish(routingKey string, body interface{}) error
}

type rabbitMQ struct {
	conn *amqp.Connection
}

func NewRabbitMQ(conn *amqp.Connection) RabbitMQ {
	return &rabbitMQ{conn: conn}
}

// routing key example i.e Order.Created.Event
func (r *rabbitMQ) Publish(routingKey string, body interface{}) error {

	ch, err := r.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to establish channel :", err)
	}

	defer ch.Close()

	err = ch.ExchangeDeclare(
		"order_exchange", // name
		"topic",          // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange :", err)
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal body to json failed :", err)
	}

	err = ch.PublishWithContext(context.TODO(),
		"order",    // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonBody,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish ", err)
	}

	return nil

}
