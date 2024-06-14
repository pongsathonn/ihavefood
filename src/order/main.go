package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/pongsathonn/food-delivery/src/order/data"
	"github.com/pongsathonn/food-delivery/src/order/event"
	pb "github.com/pongsathonn/food-delivery/src/order/genproto"
	amqp "github.com/rabbitmq/amqp091-go"
)

type orderServer struct {
	pb.UnimplementedOrderServiceServer
	db data.OrderRepo
	ev event.Eventx

	rsconn *grpc.ClientConn
}

func NewOrderServer(db data.OrderRepo, ev event.Eventx, rsconn *grpc.ClientConn) *orderServer {
	return &orderServer{
		db: db,
		ev: ev,

		rsconn: rsconn,
	}
}

func (or *orderServer) ListUserPlaceOrder(ctx context.Context, in *pb.ListUserPlaceOrderRequest) (*pb.ListUserPlaceOrderResponse, error) {

	if in.Username == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username shouldn't be empty")
	}

	resp, err := or.db.PlaceOrder(in.Username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return resp, nil

}

func (or *orderServer) PreparePlaceOrder(ctx context.Context, in *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error) {

	if in.Username == "" || in.Address == nil {
		return nil, status.Errorf(codes.InvalidArgument, "bad request ")
	}

	pm := in.PaymentMethod.String()
	if _, ok := pb.PaymentMethod_value[pm]; !ok {
		return nil, fmt.Errorf("bad request kuy")
	}

	var total int32
	for _, mn := range in.Menus {
		total += mn.Price
	}

	if in.Total != ((total + in.DeliveryFee) - in.CouponDiscount) {
		return nil, errors.New("total invalid")
	}

	// save place order
	res, err := or.db.SavePlaceOrder(in)
	if err != nil {
		return nil, fmt.Errorf("save failed %v", err)
	}

	// publish event
	routingKey := "order.placed.event"
	err = or.ev.Publish(routingKey, []byte(res.OrderId))
	if err != nil {
		return nil, fmt.Errorf("couldn't create event")
	}

	// response
	return res, nil
}

func initRabbitMQ() *amqp.Connection {
	amqpUri := os.Getenv("AMQP_URI")

	conn, err := amqp.Dial(amqpUri)
	if err != nil {
		log.Fatal(err)
	}

	return conn
}

func initMongoClient() *mongo.Client {
	mongoUri := os.Getenv("ORDER_DB_URI")

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoUri))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}

	return client
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {

	lis, err := net.Listen("tcp", os.Getenv("ORDER_URI"))
	failOnError(err, "listen failed")

	db := data.NewOrderRepo(initMongoClient())
	ev := event.NewEvent(initRabbitMQ())

	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	rsconn, err := grpc.NewClient(os.Getenv("CONPON_URI"), opt)
	failOnError(err, "failed to establish restaurant connection")

	ors := NewOrderServer(db, ev, rsconn)

	s := grpc.NewServer()
	pb.RegisterOrderServiceServer(s, ors)

	/*
		If this log not display when starting server it might be from
		- Order Database not starting
	*/
	log.Println("order service is running")

	log.Fatal(s.Serve(lis))

}
