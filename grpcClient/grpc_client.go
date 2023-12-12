// grpc/grpc_client.go
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/furkanakcadogan/rate-limit/proto"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"google.golang.org/grpc"
)

var (
	allowedRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "allowed_requests",
			Help: "Total number of allowed requests",
		},
	)
)
var (
	rejectedRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "rejected_requests",
			Help: "Total number of rejected requests",
		},
	)
)

func main() {
	Prometheusinit()
	rand.Seed(time.Now().UnixNano())
	// HTTP server
	go startHTTPServer()

	fmt.Println("Server is running on :8080")
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
	go ReadClientIDs(clientIDChan)

	for clientID := range clientIDChan {
		go CheckRateLimit(client, clientID)
	}
}

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
func Prometheusinit() {
	// Register metrics with Prometheus
	prometheus.MustRegister(allowedRequests)
	prometheus.MustRegister(rejectedRequests)
}
func startHTTPServer() {
	http.Handle("/metrics", promhttp.Handler()) // Expose the registered metrics via HTTP.
	log.Println("Metric server starting")
	err := http.ListenAndServe(":8080", nil) // Start the HTTP server.
	if err != nil {
		log.Fatalf("Failed to start metric server: %v", err)
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
		allowedRequests.Inc()
		log.Printf("prometheus inc. allowed", clientID)
	} else {
		log.Printf("Response to client ID: %s, Request Rejected.", clientID)
		rejectedRequests.Inc()
		log.Printf("prometheus inc. rejected", clientID)
	}
}
