// grpc/grpc_client.go
package grpcClient

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/furkanakcadogan/rate-limit/proto"
)

func ReadClientIDs(clientIDChan chan<- string) {
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

func CheckRateLimit(client proto.RateLimitServiceClient, clientID string) {
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
