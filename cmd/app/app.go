package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	"github.com/furkanakcadogan/rate-limit/proto" // Update this import path
)

const grpcServerAddress = "localhost:50051" // Change this to your gRPC server address

func initializeRateLimiter(redisClient *redis.Client, key string, capacity, rate int64, interval time.Duration) {
	ctx := context.Background()

	// Check if the key already exists in Redis
	if exists, _ := redisClient.Exists(ctx, key).Result(); exists == 0 {
		// Key doesn't exist, initialize the rate limiter
		result := redisClient.SetEX(ctx, key, capacity, interval)
		log.Printf("SetEX result: %v\n", result)
	}

	lastRefillKey := fmt.Sprintf("%s_last_refill", key)
	// Check if lastRefillKey already exists in Redis
	if exists, _ := redisClient.Exists(ctx, lastRefillKey).Result(); exists == 0 {
		// lastRefillKey doesn't exist, set the current time
		result := redisClient.Set(ctx, lastRefillKey, time.Now().Unix(), 0)
		log.Printf("Set result: %v\n", result)
	}
}

func refillTokens(redisClient *redis.Client, key, lastRefillKey string, capacity, rate int64, interval time.Duration) int64 {
	ctx := context.Background()

	currentTime := time.Now().Unix()
	lastRefill, _ := redisClient.Get(ctx, lastRefillKey).Int64()

	currentTokens, _ := redisClient.Get(ctx, key).Int64()

	timePassed := currentTime - lastRefill
	intervalsPassed := timePassed / int64(interval.Seconds())

	newTokens := min(capacity, currentTokens+intervalsPassed*rate)

	result := redisClient.Set(ctx, lastRefillKey, lastRefill+intervalsPassed*int64(interval.Seconds()), 0)

	//log.Printf("Set result: %v\n", result)

	result = redisClient.Set(ctx, key, newTokens, 0)
	//log.Printf("Set result: %v\n", result)
	_ = result
	return newTokens
}

func allowRequest(redisClient *redis.Client, key, lastRefillKey string, tokensRequired, capacity, rate int64, interval time.Duration) (bool, int64) {
	ctx := context.Background()

	if exists, _ := redisClient.Exists(ctx, key).Result(); exists == 0 {
		initializeRateLimiter(redisClient, key, capacity, rate, interval)
	}

	currentTokens := refillTokens(redisClient, key, lastRefillKey, capacity, rate, interval)

	if currentTokens >= tokensRequired {
		result := redisClient.DecrBy(ctx, key, tokensRequired)
		_ = result
		//log.Printf("DecrBy result: %v\n", result)
		return true, currentTokens - tokensRequired
	} else {
		return false, currentTokens
	}
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func fetchRateLimitConfigFromPostgres(db *sql.DB, clientID string) (int, time.Duration, error) {
	query := "SELECT rate_limit, refill_interval FROM rate_limits WHERE clientid = $1 LIMIT 1"
	row := db.QueryRow(query, clientID)

	var rateLimit int
	var refillInterval time.Duration

	err := row.Scan(&rateLimit, &refillInterval)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, fmt.Errorf("clientid %s not found", clientID)
		}
		return 0, 0, fmt.Errorf("error fetching data from PostgreSQL: %v", err)
	}

	return rateLimit, refillInterval, nil
}

func isRequestAllowed(redisClient *redis.Client, db *sql.DB, clientID string, tokensRequired int64) (bool, int64, error) {
	lastRefillKey := fmt.Sprintf("%s_last_refill", clientID)

	// Fetch rate limit configuration from PostgreSQL
	rateLimit, refillInterval, err := fetchRateLimitConfigFromPostgres(db, clientID)
	if err != nil {
		return false, 0, fmt.Errorf("error fetching rate limit configuration from PostgreSQL: %v", err)
	}

	// Ensure the rate limit is initialized in Redis
	initializeRateLimiter(redisClient, clientID, int64(rateLimit), int64(rateLimit), refillInterval*time.Second)

	// Check if the request is allowed
	allowed, remainingTokens := allowRequest(redisClient, clientID, lastRefillKey, tokensRequired, int64(rateLimit), int64(rateLimit), refillInterval*time.Second)

	return allowed, remainingTokens, nil
}

func main() {
	// Docker setup
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	defer func() {
		err := redisClient.Close()
		if err != nil {
			log.Printf("Error closing Redis client: %v\n", err)
		}
	}()

	// PostgreSQL setup
	pgConnStr := "user=root password=secret dbname=rate_limiting sslmode=disable"
	pgDB, err := sql.Open("postgres", pgConnStr)
	if err != nil {
		log.Printf("Error opening PostgreSQL connection: %v\n", err)
		return
	}
	defer func() {
		err := pgDB.Close()
		if err != nil {
			log.Printf("Error closing PostgreSQL connection: %v\n", err)
		}
	}()

	// Check PostgreSQL connection
	err = pgDB.Ping()
	if err != nil {
		log.Printf("Error pinging PostgreSQL: %v\n", err)
		return
	}

	clientID := "key1"

	// Initialize gRPC connection
	conn, err := grpc.Dial(grpcServerAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	// Create a gRPC client
	grpcClient := proto.NewRateLimitServiceClient(conn)

	// Example usage of isRequestAllowed function
	allowed, remainingTokens, err := isRequestAllowed(redisClient, pgDB, clientID, 1)
	_ = allowed
	_ = remainingTokens
	if err != nil {
		log.Printf("Error checking request allowance: %v\n", err)
		return
	}

	// Use gRPC client to send a request to the server
	grpcRequest := &proto.RateLimitRequest{
		ClientId:       clientID,
		TokensRequired: 1, // You can adjust this based on your use case
	}

	grpcResponse, err := grpcClient.CheckRateLimit(context.Background(), grpcRequest)
	if err != nil {
		log.Printf("Error sending gRPC request: %v\n", err)
		return
	}

	// Process the gRPC response
	if grpcResponse.GetAllowed() {
		log.Printf("%s: gRPC Request allowed. Remaining tokens: %d\n", clientID, grpcResponse.GetRemainingTokens())
	} else {
		log.Printf("%s: gRPC Request rejected. Rate limit exceeded\n", clientID)
	}
}
