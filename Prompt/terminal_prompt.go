// Prompt/terminal_prompt.go
package main

import (
	"log"
	"os"

	"github.com/furkanakcadogan/rate-limit/grpcClient" // Import the package as grpc
	"github.com/furkanakcadogan/rate-limit/proto"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	envFileLocation := "..//app.env"
	if err := godotenv.Load(envFileLocation); err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}

	grpcServerAddress := os.Getenv("GRPC_SERVER_ADDRESS")
	conn, err := grpc.Dial(grpcServerAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := proto.NewRateLimitServiceClient(conn)

	clientIDChan := make(chan string)
	go grpcClient.ReadClientIDs(clientIDChan)

	for clientID := range clientIDChan {
		go grpcClient.CheckRateLimit(client, clientID)
	}
}
