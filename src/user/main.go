package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	_ "github.com/lib/pq"
	pb "github.com/pongsathonn/food-delivery/src/user/genproto"
)

type userService struct {
	pb.UnimplementedUserServiceServer

	db *sql.DB
}

func NewUserService(db *sql.DB) *userService {
	return &userService{db: db}
}

func initPostgres() *sql.DB {

	connStr := fmt.Sprintf("postgres://donkadmin:donkpassword@localhost:5432/user_database?sslmode=disable")

	/*

		connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/user_database?sslmode=disable",
			os.Getenv("USER_POSTGRES_USER"),
			os.Getenv("USER_POSTGRES_PASS"),
			os.Getenv("USER_POSTGRES_HOST"),
			os.Getenv("USER_POSTGRES_PORT"),
		)
	*/

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	return db

}

func main() {

	db := initPostgres()
	us := NewUserService(db)

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, us)

	lis, err := net.Listen("tcp", os.Getenv("USER_SERVER_PORT"))
	if err != nil {
		log.Fatalln("failed to listen:", err)
	}

	log.Printf("user server running ")
	log.Fatalln(s.Serve(lis))

}
