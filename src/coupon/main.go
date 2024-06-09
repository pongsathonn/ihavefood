package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/pongsathonn/food-delivery/src/coupon/genproto"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/pongsathonn/food-delivery/src/coupon/handler"
)

func main() {

	lis, err := net.Listen("tcp", os.Getenv("COUPON_URI"))
	if err != nil {
		log.Println("failed to listen coupon")
	}

	s := grpc.NewServer()
	pb.RegisterCouponServiceServer(s, handler.NewCouponServer())

	log.Println("coupong server starting") //for developing
	log.Fatal(s.Serve(lis))

	//-----------------------

	opt := grpc.WithTransportCredentials(insecure.NewCredentials())

	couponPort := os.Getenv("COUPON_PORT")
	couponUri := fmt.Sprintf("localhost:%s", couponPort)

	conn, err := grpc.NewClient(couponUri, opt)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()

	client := pb.NewOrderServiceClient(conn)

	payload := pb.PlaceOrderRequest{
		Username:  "yyy",
		Email:     "mail@mail.com",
		OrderCost: 50,
	}

	_, err = client.PlaceOrder(context.TODO(), &payload)
	if err != nil {
		log.Println(err)
	}

	//----------------------------------------------
	amqpConn, err := amqp.Dial(os.Getenv("AMQP_URI"))
	failOnError(err, "failed to connect to rabbitmq")
	defer amqpConn.Close()

	ch, err := amqpConn.Channel()
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

	msgs, err := ch.Consume(
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

	for m := range msgs {
		log.Printf("%s", m.Body)
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
