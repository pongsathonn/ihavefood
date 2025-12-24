package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/grpc"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	pb "github.com/pongsathonn/ihavefood/src/authservice/genproto"
	"github.com/pongsathonn/ihavefood/src/authservice/internal"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
)

// From https://docs.cloud.google.com/sql/docs/postgres/samples/cloud-sql-postgres-databasesql-connect-connector
func connectWithConnector() (*sql.DB, error) {
	mustGetenv := func(k string) string {
		v := os.Getenv(k)
		if v == "" {
			log.Fatalf("Fatal Error in connect_connector.go: %s environment variable not set.\n", k)
		}

		return v
	}
	// Note: Saving credentials in environment variables is convenient, but not
	// secure - consider a more secure solution such as
	// Cloud Secret Manager (https://cloud.google.com/secret-manager) to help
	// keep passwords and other secrets safe.
	var (
		dbUser                 = mustGetenv("PG_USER")                  // e.g. 'my-db-user'
		dbPwd                  = mustGetenv("PG_PASS")                  // e.g. 'my-db-password'
		dbName                 = mustGetenv("AUTH_DB_NAME")             // e.g. 'my-database'
		instanceConnectionName = mustGetenv("INSTANCE_CONNECTION_NAME") // e.g. 'project:region:instance'
		usePrivate             = os.Getenv("PRIVATE_IP")
	)

	dsn := fmt.Sprintf("user=%s password=%s database=%s", dbUser, dbPwd, dbName)
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	var opts []cloudsqlconn.Option
	if usePrivate != "" {
		opts = append(opts, cloudsqlconn.WithDefaultDialOptions(cloudsqlconn.WithPrivateIP()))
	}
	// WithLazyRefresh() Option is used to perform refresh
	// when needed, rather than on a scheduled interval.
	// This is recommended for serverless environments to
	// avoid background refreshes from throttling CPU.
	opts = append(opts, cloudsqlconn.WithLazyRefresh())
	d, err := cloudsqlconn.NewDialer(context.Background(), opts...)
	if err != nil {
		return nil, err
	}
	// Use the Cloud SQL connector to handle connecting to the instance.
	// This approach does *NOT* require the Cloud SQL proxy.
	config.DialFunc = func(ctx context.Context, network, instance string) (net.Conn, error) {
		return d.Dial(ctx, instanceConnectionName)
	}
	dbURI := stdlib.RegisterConnConfig(config)
	dbPool, err := sql.Open("pgx", dbURI)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	if err := dbPool.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	return dbPool, nil
}

func startGRPCServer(s *internal.AuthService) {

	uri := fmt.Sprintf(":%s", os.Getenv("PORT"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	healthcheck := health.NewServer()
	healthgrpc.RegisterHealthServer(grpcServer, healthcheck)

	pb.RegisterAuthServiceServer(grpcServer, s)
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

func initAMQPCon() *amqp.Connection {

	uri := fmt.Sprintf("amqp://%s:%s@%s/%s",
		os.Getenv("RBMQ_USER"),
		os.Getenv("RBMQ_PASS"),
		os.Getenv("RBMQ_HOST"),
		os.Getenv("RBMQ_USER"),
	)

	maxRetries := 5
	var conn *amqp.Connection
	var err error

	for i := 1; i <= maxRetries; i++ {
		conn, err = amqp.Dial(uri)
		if err == nil {
			slog.Info("Successfully connected to AMQP")
			return conn
		}

		if i == maxRetries {
			log.Fatalf("Could not establish AMQP connection after %d attempts: %v", maxRetries, err)
		}
		time.Sleep(5 * time.Second)
	}

	log.Fatalf("Unexpected")
	return nil

}

// created by $openssl rand -base64 32
func initSigningKey() []byte {

	key := os.Getenv("JWT_SIGNING_KEY")
	if key == "" {
		log.Fatal("missing JWT_SIGNING_KEY environment variable")
	}
	return []byte(key)
}

func main() {

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {

			if a.Key == slog.SourceKey {
				source := a.Value.Any().(*slog.Source)
				source.File = filepath.Base(source.File)
			}
			return a
		},
	}))
	slog.SetDefault(logger)

	internal.SetupValidator()

	if err := initTimeZone(); err != nil {
		slog.Error("failed to init time zone", "err", err)
	}

	db, err := connectWithConnector()
	if err != nil {
		log.Fatalf("Failed to initialize PostgresDB connection: %v", err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	m, err := migrate.NewWithDatabaseInstance(
		"file:///db/migrations",
		"postgres", driver)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	auth := internal.NewAuthService(
		initSigningKey(),
		internal.NewStorage(db),
		internal.NewRabbitMQ(initAMQPCon()),
	)

	// if err := auth.CreateDemoUsers(); err != nil {
	// 	slog.Error("Failed to create Demo Users", "err", err)
	// }

	startGRPCServer(auth)
}
