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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	// .env dosyasının bir üst dizininde olduğunu belirtin
	envFileLocation := "../../app.env"

	// .env dosyasını yükle
	if err := godotenv.Load(envFileLocation); err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}

	grpcServerAddress := os.Getenv("GRPC_SERVER_ADDRESS")

	conn, err := grpc.Dial(grpcServerAddress, grpc.WithInsecure())
	if err != nil {
		log.Printf("Failed to connect to gRPC server: %v", err)
		return
	}
	defer conn.Close()

	client := proto.NewRateLimitServiceClient(conn)

	reader := bufio.NewReader(os.Stdin)

	for {
		log.Print("Enter client ID: ")
		clientID, _ := reader.ReadString('\n')
		clientID = strings.TrimSpace(clientID)

		if clientID == "" {
			log.Println("Client ID cannot be empty. Please try again.")
			continue
		}

		tokensRequired := int64(1)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel() // Ensure context cancellation.

		response, err := client.CheckRateLimit(ctx, &proto.RateLimitRequest{
			ClientId:       clientID,
			TokensRequired: tokensRequired,
		})

		if err != nil {
			st, ok := status.FromError(err)
			// Check the error message
			if ok && (st.Code() == codes.Unknown && strings.Contains(st.Message(), "client ID not found")) {
				log.Printf("Client ID: %s not found in database. Please try again.", clientID)
			} else {
				log.Printf("Error when calling client ID: %v", err)
			}
			continue
		}

		if response.Allowed {
			log.Printf("Response to client ID: %s, allowed: %t, remaining tokens: %d\n", clientID, response.Allowed, response.RemainingTokens)
		} else {
			log.Printf("Response to client ID: %s, Request Rejected.", clientID)
		}
	}
}
