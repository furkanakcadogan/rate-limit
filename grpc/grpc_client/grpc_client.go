package grpc_client

import (
	"log"

	"github.com/furkanakcadogan/rate-limit/cmd/app" // Import the app package
	"github.com/furkanakcadogan/rate-limit/proto"

	"google.golang.org/grpc"
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

	// Call the IsRequestAllowed function in cmd/app/app.go
	allowed, remainingTokens, err := app.IsRequestAllowed(client, "someClientId", 5)
	if err != nil {
		log.Fatalf("Error checking request allowance: %v", err)
	}

	// Print the response
	if allowed {
		log.Printf("gRPC Request allowed. Remaining tokens: %d\n", remainingTokens)
	} else {
		log.Printf("gRPC Request rejected. Rate limit exceeded\n")
	}
}
