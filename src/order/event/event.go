package event

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Eventx interface {
	Publish(routingKey string, body []byte) error
}

type eventx struct {
	conn *amqp.Connection
}

func NewEvent(conn *amqp.Connection) Eventx {
	return &eventx{conn: conn}
}

// routing key example i.e Order.Created.Event
func (e *eventx) Publish(routingKey string, body []byte) error {

	ch, err := e.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to establish channel :", err)
	}
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
	if err != nil {
		return fmt.Errorf("failed to declare exchange :", err)
	}

	err = ch.PublishWithContext(context.TODO(),
		"order",    // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish ", err)
	}

	return nil

}
