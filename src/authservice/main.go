package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/pongsathonn/ihavefood/src/authservice/internal"

	_ "github.com/lib/pq"
	pb "github.com/pongsathonn/ihavefood/src/authservice/genproto"
)

func initPostgres() (*sql.DB, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user := os.Getenv("AUTH_DB_USER")
	pass := os.Getenv("AUTH_DB_PASS")
	host := os.Getenv("AUTH_DB_HOST")
	dbName := os.Getenv("AUTH_DB_NAME")

	db, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		user,
		pass,
		host,
		dbName,
	))
	if err != nil {
		return nil, err
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	slog.Info("Database initialized successfully", "host", host)
	return db, nil
}

// startGRPCServer sets up and starts the gRPC server
// func startGRPCServer(s *internal.AuthService) {
func startGRPCServer(s *internal.AuthService) {

	uri := fmt.Sprintf(":%s", os.Getenv("AUTH_SERVER_PORT"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, s)

	slog.Info("AuthService initialized successfully",
		"port", os.Getenv("AUTH_SERVER_PORT"),
	)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

func initTimeZone() error {
	l, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		return err
	}

	time.Local = l
	return nil
}

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
	slog.SetDefault(logger)

	if err := initTimeZone(); err != nil {
		slog.Error("failed to init time zone", "err", err)
	}

	if err := internal.InitSigningKey(); err != nil {
		slog.Error("failed to init jwt signing key", "err", err)
	}

	db, err := initPostgres()
	if err != nil {
		log.Fatalf("Failed to initialize PostgresDB connection: %v", err)
	}
	storage := internal.NewStorage(db)

	if err := internal.CreateSuperAdmin(storage); err != nil {
		slog.Error("Failed to create admin user", "err", err)
	}

	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	customers, err := grpc.NewClient(os.Getenv("CUSTOMER_URI"), opt)
	if err != nil {
		log.Fatalf("Failed to initialize CustomerService connection: %v", err)
	}

	deliv, err := grpc.NewClient(os.Getenv("DELIVERY_URI"), opt)
	if err != nil {
		log.Fatalf("Failed to initialize DeliveryService connection: %v", err)
	}

	merchant, err := grpc.NewClient(os.Getenv("MERCHANT_URI"), opt)
	if err != nil {
		log.Fatalf("Failed to initialize MerchantService connection: %v", err)
	}
	slog.Info("Downstream gRPC channels created successfully")

	startGRPCServer(internal.NewAuthService(&internal.AuthCfg{
		Store:          storage,
		CustomerClient: pb.NewCustomerServiceClient(customers),
		DeliveryClient: pb.NewDeliveryServiceClient(deliv),
		MerchantClient: pb.NewMerchantServiceClient(merchant),
	}))
}
