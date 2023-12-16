package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
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

	// Load .env file
	envFileLocation := "..//app.env"
	if err := godotenv.Load(envFileLocation); err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}

	// Read HTTP port from environment variable
	httpPort := os.Getenv("GO_CLIENT_HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080" // Default port if not specified
	}

	// HTTP server
	go startHTTPServer(httpPort)

	grpcServerAddress := os.Getenv("GRPC_SERVER_ADDRESS")
	conn, err := grpc.Dial(grpcServerAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := proto.NewRateLimitServiceClient(conn)

	http.HandleFunc("/check-rate-limit", func(w http.ResponseWriter, r *http.Request) {
		handleCheckRateLimit(client, w, r)
	})

	// Print server running message with the actual port
	fmt.Printf("Server is running on :%s\n", httpPort)

	// Blocking the main goroutine while other goroutines are running
	select {}
}
func Prometheusinit() {
	// Register metrics with Prometheus
	prometheus.MustRegister(allowedRequests)
	prometheus.MustRegister(rejectedRequests)
}

func startHTTPServer(port string) {
	http.Handle("/metrics", promhttp.Handler()) // Expose the registered metrics via HTTP.
	log.Printf("Metric server starting on port %s\n", port)
	err := http.ListenAndServe(":"+port, nil) // Start the HTTP server.
	if err != nil {
		log.Fatalf("Failed to start metric server: %v", err)
	}
}
func handleCheckRateLimit(client proto.RateLimitServiceClient, w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		ClientID string `json:"client_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}
	response, err := CheckRateLimit(client, request.ClientID)
	if err != nil {
		http.Error(w, "Error checking rate limit", http.StatusInternalServerError)
		return
	}

	// Yanıtı JSON olarak hazırla ve gönder
	var message string
	if response.Allowed {
		message = "Request Allowed"
		w.WriteHeader(http.StatusOK) // 200 OK
	} else {
		message = "Request Rejected due to Rate Limit"
		w.WriteHeader(http.StatusTooManyRequests) // 429 Too Many Requests
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":         message,
		"client_id":       request.ClientID,
		"remainingTokens": response.RemainingTokens,
	})
}

func CheckRateLimit(client proto.RateLimitServiceClient, clientID string) (*proto.RateLimitResponse, error) {
	tokensRequired := int64(1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	response, err := client.CheckRateLimit(ctx, &proto.RateLimitRequest{
		ClientId:       clientID,
		TokensRequired: tokensRequired,
	})
	if err != nil {
		return nil, err
	}

	if response.Allowed {
		allowedRequests.Inc()
	} else {
		rejectedRequests.Inc()
		// Burada, `response` nesnesi hala geçerli ve `response.Allowed` false olduğunda
		// HTTP yanıtı olarak kullanılabilir.
	}

	return response, nil
}
