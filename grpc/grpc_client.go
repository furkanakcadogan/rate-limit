package grpc

import (
	"context"
	"log"

	"google.golang.org/grpc"

	"github.com/furkanakcadogan/rate-limit/proto"
)

func grpc_client() {
	// Create a gRPC connection to the server
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create a gRPC client
	client := proto.NewRateLimitServiceClient(conn)

	// Send a CheckRateLimit request
	request := &proto.RateLimitRequest{
		ClientId:       "someClientId",
		TokensRequired: 5,
	}

	response, err := client.CheckRateLimit(context.Background(), request)
	if err != nil {
		log.Fatalf("Error sending CheckRateLimit request: %v", err)
	}

	// Print the response
	log.Printf("Response: %+v", response)
}
