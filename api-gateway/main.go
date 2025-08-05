package main

import (
	"github.com/pongsathonn/ihavefood/api-gateway/gateway"
	"github.com/pongsathonn/ihavefood/api-gateway/server"
	"log"
)

func main() {

	gwmux := gateway.New().SetupMux()
	if err := server.Run(gwmux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
