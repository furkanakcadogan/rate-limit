package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"google.golang.org/grpc"

	"github.com/furkanakcadogan/rate-limit/proto"
)

var (
	grpcServerAddress string
	redisAddress      string
	pgConnStr         string
)

var (
	isDynamicRateLimiting bool
)

var (
	currentCPUUsage    float64
	currentMemoryUsage float64
)

type RateLimitServer struct {
	proto.UnimplementedRateLimitServiceServer
	redisClient *redis.Client
	db          *sql.DB
}

func NewRateLimitServer(redisClient *redis.Client, db *sql.DB) *RateLimitServer {
	return &RateLimitServer{
		redisClient: redisClient,
		db:          db,
	}
}

func (s *RateLimitServer) CheckRateLimit(ctx context.Context, req *proto.RateLimitRequest) (*proto.RateLimitResponse, error) {
	clientID := req.ClientId
	tokensRequired := req.TokensRequired

	log.Printf("CheckRateLimit called with clientId: %s, tokensRequired: %d\n", clientID, tokensRequired)

	allowed, remainingTokens, err := isRequestAllowed(s.redisClient, s.db, clientID, tokensRequired, isDynamicRateLimiting)

	if err != nil {
		log.Printf("Error in isRequestAllowed for clientId: %s, error: %v\n", clientID, err)
		return nil, err
	}

	if allowed {
		log.Printf("Response to clientId: %s, allowed: %t, remainingTokens: %d\n", clientID, allowed, remainingTokens)
	} else {
		log.Printf("Response to clientId: %s, Request Rejected.", clientID)
	}

	return &proto.RateLimitResponse{
		Allowed:         allowed,
		RemainingTokens: remainingTokens,
	}, nil
}

func initializeRateLimiter(redisClient *redis.Client, key string, capacity, rate int64, interval time.Duration) {
	ctx := context.Background()

	if exists, err := redisClient.Exists(ctx, key).Result(); err != nil || exists == 0 {
		if err != nil {
			log.Printf("Error checking if key %s exists in Redis: %v\n", key, err)
		}
		_, err := redisClient.SetEX(ctx, key, capacity, interval).Result()
		if err != nil {
			log.Printf("Error initializing rate limiter in Redis for key %s: %v\n", key, err)
		} else {
			log.Printf("Initialized rate limiter in Redis for key %s with capacity %d\n", key, capacity)
		}
	}

	lastRefillKey := fmt.Sprintf("%s_last_refill", key)
	if exists, err := redisClient.Exists(ctx, lastRefillKey).Result(); err != nil || exists == 0 {
		if err != nil {
			log.Printf("Error checking if last refill key %s exists in Redis: %v\n", lastRefillKey, err)
		}
		_, err := redisClient.Set(ctx, lastRefillKey, time.Now().Unix(), 0).Result()
		if err != nil {
			log.Printf("Error setting last refill time in Redis for key %s: %v\n", lastRefillKey, err)
		} else {
			log.Printf("Set last refill time in Redis for key %s\n", lastRefillKey)
		}
	}
}

func refillTokens(redisClient *redis.Client, key, lastRefillKey string, capacity, rate int64, interval time.Duration) int64 {
	ctx := context.Background()

	currentTime := time.Now().Unix()
	lastRefill, err := redisClient.Get(ctx, lastRefillKey).Int64()
	if err != nil {
		log.Printf("Error getting last refill time from Redis: %v\n", err)
		return 0
	}

	currentTokens, err := redisClient.Get(ctx, key).Int64()
	if err != nil {
		log.Printf("Error getting current tokens from Redis: %v\n", err)
		return 0
	}

	timePassed := currentTime - lastRefill
	intervalsPassed := timePassed / int64(interval.Seconds())
	newTokens := min(capacity, currentTokens+intervalsPassed*rate)

	_, err = redisClient.Set(ctx, lastRefillKey, currentTime, 0).Result()
	if err != nil {
		log.Printf("Error updating last refill time in Redis: %v\n", err)
	}

	_, err = redisClient.Set(ctx, key, newTokens, 0).Result()
	if err != nil {
		log.Printf("Error updating token count in Redis: %v\n", err)
	}

	return newTokens
}

func allowRequest(redisClient *redis.Client, key, lastRefillKey string, tokensRequired, capacity, rate int64, interval time.Duration) (bool, int64) {
	ctx := context.Background()

	if exists, err := redisClient.Exists(ctx, key).Result(); err != nil || exists == 0 {
		initializeRateLimiter(redisClient, key, capacity, rate, interval)
	}

	currentTokens := refillTokens(redisClient, key, lastRefillKey, capacity, rate, interval)

	if currentTokens >= tokensRequired {
		_, err := redisClient.DecrBy(ctx, key, tokensRequired).Result()
		if err != nil {
			log.Printf("Error decrementing tokens in Redis: %v\n", err)
		}
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

func fetchRateLimitConfigFromPostgres(db *sql.DB, clientID string) (int64, time.Duration, error) {
	query := "SELECT rate_limit, refill_interval FROM ratelimitingdb WHERE clientid = $1 LIMIT 1"
	row := db.QueryRow(query, clientID)

	var rateLimit int64
	var refillInterval int
	err := row.Scan(&rateLimit, &refillInterval)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("client ID %s not found in database\n", clientID)
			return 0, 0, fmt.Errorf("client ID %s not found", clientID)
		}
		log.Printf("error fetching data from PostgreSQL for client ID %s: %v\n", clientID, err)
		return 0, 0, fmt.Errorf("error fetching data from PostgreSQL: %v", err)
	}

	log.Printf("Fetched rateLimit: %d, refillInterval: %d for client ID %s\n", rateLimit, refillInterval, clientID)
	return rateLimit, time.Duration(refillInterval) * time.Second, nil
}

func isRequestAllowed(redisClient *redis.Client, db *sql.DB, clientID string, tokensRequired int64, isDynamicRateLimiting bool) (bool, int64, error) {
	lastRefillKey := fmt.Sprintf("%s_last_refill", clientID)

	rateLimit, refillInterval, err := fetchRateLimitConfigFromPostgres(db, clientID)
	if err != nil {
		return false, 0, err
	}

	if isDynamicRateLimiting {
		rateLimit = calculateNewLimits(rateLimit, currentCPUUsage, currentMemoryUsage, true)
	}

	initializeRateLimiter(redisClient, clientID, rateLimit, rateLimit, refillInterval)

	allowed, remainingTokens := allowRequest(redisClient, clientID, lastRefillKey, tokensRequired, rateLimit, rateLimit, refillInterval)

	return allowed, remainingTokens, nil
}

func getSystemLoad() (float64, float64, error) {
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		log.Printf("Error retrieving CPU usage: %v\n", err)
		return 0, 0, err
	}

	memStat, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("Error retrieving memory usage: %v\n", err)
		return 0, 0, err
	}

	cpuUsage := cpuPercent[0] / 100
	memoryUsage := memStat.UsedPercent / 100

	return cpuUsage, memoryUsage, nil
}

func calculateNewLimits(originalRateLimit int64, cpuUsage float64, memoryUsage float64, isDynamicRateLimitting bool) int64 {
	if !isDynamicRateLimitting {
		return originalRateLimit
	}

	var multiplier float64

	switch {
	case cpuUsage > 0.97 || memoryUsage > 0.97:
		multiplier = 0.1
	case cpuUsage > 0.95 || memoryUsage > 0.92:
		multiplier = 0.2
	case cpuUsage > 0.90 || memoryUsage > 0.90:
		multiplier = 0.3
	case cpuUsage > 0.85 || memoryUsage > 0.89:
		multiplier = 0.4
	case cpuUsage > 0.80 || memoryUsage > 0.87:
		multiplier = 0.7
	case cpuUsage > 0.75 || memoryUsage > 0.85:
		multiplier = 0.8
	case cpuUsage > 0.70 || memoryUsage > 0.80:
		multiplier = 0.9
	default:
		multiplier = 1.0
	}

	newRateLimit := float64(originalRateLimit) * multiplier
	return int64(newRateLimit)
}

func startGRPCServer(grpcServerAddress string, redisClient *redis.Client, db *sql.DB) {
	lis, err := net.Listen("tcp", grpcServerAddress)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterRateLimitServiceServer(grpcServer, NewRateLimitServer(redisClient, db))

	log.Printf("Starting gRPC server on %s", grpcServerAddress)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
func pingServer(server string) bool {
	timeout := 2 * time.Second
	_, err := net.DialTimeout("tcp", server, timeout)
	if err != nil {
		log.Printf("Failed to connect to server %s: %v\n", server, err)
		return false
	}
	log.Printf("Successfully pinged server %s\n", server)
	return true
}

func pingPostgresDB(ctx context.Context, db *sql.DB) {
	if err := db.PingContext(ctx); err != nil {
		log.Printf("Failed to ping PostgreSQL database: %v\n", err)
	} else {
		log.Println("Successfully pinged PostgreSQL database")
	}
}

func main() {

	isDynamicRateLimiting = true

	if isDynamicRateLimiting {
		log.Println("Dynamic Rate Limiting is ENABLED")
	} else {
		log.Println("Dynamic Rate Limiting is DISABLED")
	}
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			cpuUsage, memoryUsage, err := getSystemLoad()
			if err != nil {
				log.Printf("Error getting system load: %v\n", err)
				continue
			}
			currentCPUUsage = cpuUsage
			currentMemoryUsage = memoryUsage
		}
	}()

	grpcServerAddress = ":50051"

	redisAddress = "rate-limit-redis.jiq88u.ng.0001.eun1.cache.amazonaws.com:6379"
	pgConnStr = "postgresql://root:zNTcuDav4wU8EnZ4Wnp3@rate-limit.c5ntee3dn9xx.eu-north-1.rds.amazonaws.com:5432/ratelimitingdb?sslmode=require"

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddress,
		DB:   0,
	})
	defer redisClient.Close()

	db, err := sql.Open("postgres", pgConnStr)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	pingServer(redisAddress)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pingPostgresDB(ctx, db)

	go startGRPCServer(grpcServerAddress, redisClient, db)

	select {}
}
