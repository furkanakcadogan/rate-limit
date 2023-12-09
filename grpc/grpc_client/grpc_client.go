package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/furkanakcadogan/rate-limit/proto"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	envFileLocation := "../../app.env"
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
	go readClientIDs(clientIDChan)

	for clientID := range clientIDChan {
		go checkRateLimit(client, clientID)
	}
}

func readClientIDs(clientIDChan chan<- string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		log.Print("Enter client ID: ")
		clientID, _ := reader.ReadString('\n')
		clientID = strings.TrimSpace(clientID)

		if clientID != "" {
			clientIDChan <- clientID
		}
	}
}

func checkRateLimit(client proto.RateLimitServiceClient, clientID string) {
	tokensRequired := int64(1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	response, err := client.CheckRateLimit(ctx, &proto.RateLimitRequest{
		ClientId:       clientID,
		TokensRequired: tokensRequired,
	})
	if err != nil {
		log.Printf("Error when calling client ID: %v", err)
		return
	}

	if response.Allowed {
		log.Printf("Response to client ID: %s, allowed: %t, remaining tokens: %d\n", clientID, response.Allowed, response.RemainingTokens)
	} else {
		log.Printf("Response to client ID: %s, Request Rejected.", clientID)
	}
}
