package main

import (
	"context"
	"log"

	pb "github.com/pongsathonn/food-delivery/src/coupon/genproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient("localhost:4010", opt)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()

	client := pb.NewOrderServiceClient(conn)

	payload := pb.PlaceOrderRequest{
		Username:  "xxx",
		Email:     "mail@mail.com",
		OrderCost: 50,
	}

	res, err := client.PlaceOrder(context.TODO(), &payload)
	if err != nil {
		log.Println(err)
	}

	log.Println(res.OrderId, "\n", res.OrderTrackingId)

}
