package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/furkanakcadogan/rate-limit/db"
	"github.com/go-redis/redis/v8"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/joho/godotenv"
)

var (
	redisAddress string
	redisClient  *redis.Client
	queries      *db.Queries
)

func main() {
	envFileLocation := "../app.env"

	// Load .env file
	if err := godotenv.Load(envFileLocation); err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}

	// Setup database connection
	pgConnStr := "postgresql://root:secret@localhost:5432/ratelimitingdb?sslmode=disable"
	conn, err := sql.Open("pgx", pgConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer conn.Close()

	queries = db.New(conn)

	// Get Redis address from environment variable
	redisAddress = os.Getenv("REDIS_ADDRESS")
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: "",
		DB:       0,
	})

	// Setup HTTP route handlers
	http.HandleFunc("/refresh/id", handleRefreshID)
	http.HandleFunc("/refresh/all", handleRefreshAll)
	httpPort := os.Getenv("REDIS_REFRESHER_HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080" // Default port if not specified
	}

	// Start HTTP server
	fmt.Printf("Starting server at port %s\n", httpPort)
	log.Fatal(http.ListenAndServe(":"+httpPort, nil))

}

func handleRefreshID(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	var request struct {
		ClientID string `json:"client_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err := refreshID(context.Background(), request.ClientID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "ID not found in the database")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Refreshed Redis Cache for ID: " + request.ClientID})
}

func handleRefreshAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}

	refreshAll(context.Background())
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "All Redis cache refreshed successfully"})
}

func refreshID(ctx context.Context, clientID string) error {
	rateLimitData, err := queries.GetRateLimit(ctx, clientID)
	if err != nil {
		log.Printf("Error retrieving rate limit data for Client ID %s: %v\n", clientID, err)
		return err
	}

	key := clientID
	lastRefillKey := fmt.Sprintf("%s_last_refill", key)

	_, err = redisClient.SetEX(ctx, key, int64(rateLimitData.RateLimit), time.Duration(rateLimitData.RefillInterval)*time.Second).Result()
	if err != nil {
		log.Printf("Error resetting token count for Client ID %s: %v\n", clientID, err)
		return err
	}

	_, err = redisClient.Set(ctx, lastRefillKey, time.Now().Unix(), 0).Result()
	if err != nil {
		log.Printf("Error resetting last refill time for Client ID %s: %v\n", clientID, err)
		return err
	}

	log.Printf("Rate limiter for Client ID %s has been refreshed successfully.\n", clientID)
	return nil
}

func refreshAll(ctx context.Context) {
	err := redisClient.FlushDB(ctx).Err()
	if err != nil {
		log.Println("Error refreshing all Redis cache:", err)
		return
	}

	log.Println("All Redis cache refreshed successfully.")
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
