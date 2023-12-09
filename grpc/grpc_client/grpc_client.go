package main

import (
	"context"
	"log"
	"time"

	"github.com/furkanakcadogan/rate-limit/proto" // Bu import yolu projenize göre güncellenmelidir.
	"google.golang.org/grpc"
)

const grpcServerAddress = "localhost:50051"

func main() {
	conn, err := grpc.Dial(grpcServerAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := proto.NewRateLimitServiceClient(conn)

	clientID := "key1"         // Burada test etmek istediğiniz client ID'yi kullanın.
	tokensRequired := int64(1) // Gerekli token sayısı.

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	response, err := client.CheckRateLimit(ctx, &proto.RateLimitRequest{
		ClientId:       clientID,
		TokensRequired: tokensRequired,
	})
	if err != nil {
		log.Fatalf("Error when calling CheckRateLimit: %v", err)
	}

	if response.Allowed {
		log.Printf("Response to clientId: %s, allowed: %t, remainingTokens: %d\n", clientID, response.Allowed, response.RemainingTokens)
	} else {
		log.Printf("Response to clientId: %s, Request Rejected.", clientID)
	}
}
