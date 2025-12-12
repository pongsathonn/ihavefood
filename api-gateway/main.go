package main

import (
	"github.com/pongsathonn/ihavefood/api-gateway/server"
	"log"
)

func main() {

	if err := server.Run(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

}
