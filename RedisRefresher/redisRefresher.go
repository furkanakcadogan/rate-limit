package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/furkanakcadogan/rate-limit/db"
	"github.com/go-redis/redis/v8"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/joho/godotenv"
)

var redisAddress string // Declare the redisAddress variable

func main() {
	envFileLocation := "../app.env"

	// Load .env file
	if err := godotenv.Load(envFileLocation); err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}
	pgConnStr := "postgresql://root:secret@localhost:5432/ratelimitingdb?sslmode=disable"
	conn, err := sql.Open("pgx", pgConnStr)
	if err != nil {
		log.Fatalf("Veritabanına bağlanılamadı: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	// Get Redis address from environment variable
	redisAddress = os.Getenv("REDIS_ADDRESS")
	reader := bufio.NewReader(os.Stdin)

	// Initialize database connection and db.Queries (Add your database connection logic here)
	// dbConn := ...
	// queries := db.New(dbConn)

	// Ask user for action choice
	fmt.Println("Do you want to refresh a specific ID or all Redis cache? (id/all)")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddress, // Use the redisAddress variable
		Password: "",           // No password
		DB:       0,            // Use default DB
	})
	ctx := context.Background()

	// Handle user choice
	switch choice {
	case "id":
		// Refresh specific ID
		fmt.Println("Enter the ID to refresh:")
		id, _ := reader.ReadString('\n')
		id = strings.TrimSpace(id)

		// Call the refreshID function with the correct arguments
		refreshID(ctx, redisClient, queries, id) // Add the queries argument

	case "all":
		// Refresh all Redis cache
		fmt.Println("Refreshing all Redis cache...")
		refreshAll(ctx, redisClient)

	default:
		fmt.Println("Invalid choice. Please enter 'id' or 'all'.")
	}
}

func refreshID(ctx context.Context, redisClient *redis.Client, queries *db.Queries, id string) {
	// Veritabanından kapasite ve hız sınırı değerlerini alın
	rateLimitData, err := queries.GetRateLimit(ctx, id)
	if err != nil {
		log.Printf("Error retrieving rate limit data for ID %s: %v\n", id, err)
		return
	}

	// Anahtarlar
	key := id // Veya belirli bir format kullanın, örneğin: fmt.Sprintf("rate_limit_%s", id)
	lastRefillKey := fmt.Sprintf("%s_last_refill", key)

	// Mevcut token sayısını ve son dolum zamanını sıfırlayın
	_, err = redisClient.SetEX(ctx, key, int64(rateLimitData.RateLimit), time.Duration(rateLimitData.RefillInterval)*time.Second).Result()
	if err != nil {
		log.Printf("Error resetting token count for ID %s: %v\n", id, err)
		return
	}

	_, err = redisClient.Set(ctx, lastRefillKey, time.Now().Unix(), 0).Result()
	if err != nil {
		log.Printf("Error resetting last refill time for ID %s: %v\n", id, err)
		return
	}

	log.Printf("Rate limiter for ID %s has been refreshed successfully.\n", id)
}

func refreshAll(ctx context.Context, redisClient *redis.Client) {
	err := redisClient.FlushDB(ctx).Err()
	if err != nil {
		fmt.Println("Error refreshing all Redis cache:", err)
	} else {
		fmt.Println("All Redis cache refreshed successfully.")
	}
}
