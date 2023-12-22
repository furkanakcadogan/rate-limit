package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/furkanakcadogan/rate-limit/db"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/lib/pq"
)

type CreateNewUserParams struct {
	ClientID       string `json:"clientId"`
	RateLimit      int32  `json:"rateLimit"`
	RefillInterval int32  `json:"refillInterval"`
}

type HTTPHandler struct {
	queries *db.Queries
}

func NewHTTPHandler(queries *db.Queries) *HTTPHandler {
	return &HTTPHandler{queries: queries}
}

func handleHTTPMethod(w http.ResponseWriter, r *http.Request, allowedMethod string, handlerFunc http.HandlerFunc) {
	if r.Method != allowedMethod {
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
		return
	}
	handlerFunc(w, r)
}

func (h *HTTPHandler) InsertNewClientHandler(w http.ResponseWriter, r *http.Request) {
	var params CreateNewUserParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	message, err := InsertNewClient(h.queries, params.ClientID, params.RateLimit, params.RefillInterval)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func InsertNewClient(queries *db.Queries, clientID string, rateLimit int32, refillInterval int32) (string, error) {
	_, err := queries.GetRateLimit(context.Background(), clientID)
	if err == nil {
		msg := fmt.Sprintf("ClientID %s already exists in the database. Skipping duplicate entry.", clientID)
		log.Print(msg)
		return msg, nil
	} else if err != sql.ErrNoRows {
		return "", err
	}

	createParams := db.CreateRateLimitParams{
		Clientid:       clientID,
		RateLimit:      rateLimit,
		RefillInterval: refillInterval,
	}

	_, err = queries.CreateRateLimit(context.Background(), createParams)
	if err != nil {
		return "", err
	}

	msg := fmt.Sprintf("New client with ID %s successfully inserted.", clientID)
	return msg, nil
}

func (h *HTTPHandler) DeleteClientHandler(w http.ResponseWriter, r *http.Request) {
	handleHTTPMethod(w, r, "POST", func(w http.ResponseWriter, r *http.Request) {
		var params DeleteClientParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		message, err := DeleteClient(h.queries, params.ClientID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": message})
	})
}

type DeleteClientParams struct {
	ClientID string `json:"clientId"`
}

func DeleteClient(queries *db.Queries, clientID string) (string, error) {
	_, err := queries.GetRateLimit(context.Background(), clientID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Sprintf("Client %s does not exist in the database.", clientID), nil
		}
		return "", err
	}

	err = queries.DeleteRateLimit(context.Background(), clientID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Client %s successfully deleted.", clientID), nil
}

type ListClientParams struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func (h *HTTPHandler) ListClientIDRecordsHandler(w http.ResponseWriter, r *http.Request) {
	handleHTTPMethod(w, r, "POST", func(w http.ResponseWriter, r *http.Request) {
		var params ListClientParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if params.Limit <= 0 {
			params.Limit = 100
		}
		if params.Offset < 0 {
			params.Offset = 0
		}

		dbParams := db.ListRateLimitsParams{
			Limit:  int32(params.Limit),
			Offset: int32(params.Offset),
		}

		result, err := h.queries.ListRateLimits(context.Background(), dbParams)
		if err != nil {
			log.Printf("ListRateLimits query failed: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	})
}

type GenerateRandomClientParams struct {
	NumRecords int `json:"numRecords"`
}

func (h *HTTPHandler) GenerateRandomClientIDsHandler(w http.ResponseWriter, r *http.Request) {
	handleHTTPMethod(w, r, "POST", func(w http.ResponseWriter, r *http.Request) {
		var params GenerateRandomClientParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if params.NumRecords <= 0 {
			http.Error(w, "Number of records must be positive", http.StatusBadRequest)
			return
		}

		message, err := GenerateRandomClientIDs(h.queries, params.NumRecords)
		if err != nil {
			http.Error(w, "Cannot Generate All Users There Are Some Duplicates", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": message})
	})
}

func GenerateRandomClientIDs(queries *db.Queries, numRecords int) (string, error) {
	rand.Seed(time.Now().UnixNano())

	successfulRecords := 0

	for i := 0; i < numRecords; i++ {
		rateLimit := int32(rand.Intn(21))
		refillInterval := int32(rand.Intn(51) + 10)
		clientID := fmt.Sprintf("Client%d", rand.Intn(101))

		createParams := db.CreateRateLimitParams{
			Clientid:       clientID,
			RateLimit:      rateLimit,
			RefillInterval: refillInterval,
		}

		_, err := queries.CreateRateLimit(context.Background(), createParams)
		if err != nil {
			if isDuplicateKeyError(err) {
				log.Printf("Skipping record creation for %s due to duplicate key violation.", createParams.Clientid)
				continue
			}
			log.Printf("Error creating rate limit: %v", err)
		} else {
			successfulRecords++

			fetchedRateLimit, err := queries.GetRateLimit(context.Background(), createParams.Clientid)
			if err != nil {
				log.Printf("Error fetching rate limit: %v", err)
			} else if fetchedRateLimit.Clientid != createParams.Clientid ||
				fetchedRateLimit.RateLimit != rateLimit ||
				fetchedRateLimit.RefillInterval != refillInterval {
				log.Printf("Fetched rate limit does not match the created one.")
			}
		}
	}

	return fmt.Sprintf("%d random clients generated successfully, The other random values already in the database ", successfulRecords), nil
}

func isDuplicateKeyError(err error) bool {
	if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
		return true
	}
	return false
}

type UpdateClientParametersParams struct {
	ClientID          string `json:"clientId"`
	NewRateLimit      int32  `json:"newRateLimit"`
	NewRefillInterval int32  `json:"newRefillInterval"`
}

type RateLimitParams struct {
	ClientID       string `json:"clientId"`
	RateLimit      int32  `json:"rateLimit"`
	RefillInterval int32  `json:"refillInterval"`
}

func UpdateClientParameters(queries *db.Queries, clientID string, newRateLimit int32, newRefillInterval int32) (RateLimitParams, error) {
	currentParams, err := queries.GetRateLimit(context.Background(), clientID)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("Client %s not found in the database. Continuing...\n", clientID)
			return RateLimitParams{}, err
		} else {
			log.Printf("GetRateLimit query failed: %v", err)
			return RateLimitParams{}, err
		}
	}

	fmt.Printf("Current Rate Limit for %s: %d\n", clientID, currentParams.RateLimit)
	fmt.Printf("Current Refill Interval for %s: %d\n", clientID, currentParams.RefillInterval)

	updateParams := db.UpdateRateLimitParams{
		Clientid:       clientID,
		RateLimit:      newRateLimit,
		RefillInterval: newRefillInterval,
	}

	err = queries.UpdateRateLimit(context.Background(), updateParams)

	if err != nil {
		log.Printf("UpdateRateLimit query failed: %v", err)
		return RateLimitParams{}, err
	}

	updatedParams, err := queries.GetRateLimit(context.Background(), clientID)

	if err != nil {
		log.Printf("GetRateLimit query failed: %v", err)
		return RateLimitParams{}, err
	}

	fmt.Printf("Updated Rate Limit for %s: %d\n", clientID, updatedParams.RateLimit)
	fmt.Printf("Updated Refill Interval for %s: %d\n", clientID, updatedParams.RefillInterval)

	updatedRateLimitParams := RateLimitParams{
		ClientID:       updatedParams.Clientid,
		RateLimit:      updatedParams.RateLimit,
		RefillInterval: updatedParams.RefillInterval,
	}

	return updatedRateLimitParams, nil
}

func (h *HTTPHandler) UpdateClientParametersHandler(w http.ResponseWriter, r *http.Request) {
	handleHTTPMethod(w, r, "POST", func(w http.ResponseWriter, r *http.Request) {
		var params UpdateClientParametersParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		updatedParams, err := UpdateClientParameters(h.queries, params.ClientID, params.NewRateLimit, params.NewRefillInterval)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updatedParams)
	})
}

func DeleteAllClients(queries *db.Queries) error {
	fmt.Println("Deleting all rate limits from the database...")

	err := queries.DeleteAllRateLimits(context.Background())
	if err != nil {
		log.Printf("Error deleting all rate limits: %v", err)
		return err
	}
	fmt.Println("All rate limits deleted successfully.")
	return nil
}

func (h *HTTPHandler) DeleteAllClientsHandler(w http.ResponseWriter, r *http.Request) {
	handleHTTPMethod(w, r, "POST", func(w http.ResponseWriter, r *http.Request) {
		err := DeleteAllClients(h.queries)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "All rate limits deleted successfully."})
	})
}

func main() {
	// Setup database connection using environment variable
	pgConnStr := "postgresql://root:zNTcuDav4wU8EnZ4Wnp3@rate-limit.c5ntee3dn9xx.eu-north-1.rds.amazonaws.com:5432/ratelimitingdb?sslmode=require"
	if pgConnStr == "" {
		log.Fatalf("Database source not provided in the environment")
	}
	conn, err := sql.Open("pgx", pgConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer conn.Close()
	// Ping the database to check for a successful connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = conn.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping the database: %v", err)
	}

	queries := db.New(conn)
	handler := NewHTTPHandler(queries)

	// HTTP handlers setup
	http.HandleFunc("/insert", func(w http.ResponseWriter, r *http.Request) {
		handleHTTPMethod(w, r, "POST", handler.InsertNewClientHandler)
	})

	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		handleHTTPMethod(w, r, "POST", handler.DeleteClientHandler)
	})

	http.HandleFunc("/list-clients", func(w http.ResponseWriter, r *http.Request) {
		handleHTTPMethod(w, r, "POST", handler.ListClientIDRecordsHandler)
	})

	http.HandleFunc("/generate-random-clients", func(w http.ResponseWriter, r *http.Request) {
		handleHTTPMethod(w, r, "POST", handler.GenerateRandomClientIDsHandler)
	})

	http.HandleFunc("/update-client-parameters", func(w http.ResponseWriter, r *http.Request) {
		handleHTTPMethod(w, r, "POST", handler.UpdateClientParametersHandler)
	})

	http.HandleFunc("/delete-all-clients", func(w http.ResponseWriter, r *http.Request) {
		handleHTTPMethod(w, r, "POST", handler.DeleteAllClientsHandler)
	})

	// Setup HTTP server port
	httpPort := "8082"
	if httpPort == "" {
		httpPort = "8082"
	}

	// Start HTTP server
	fmt.Printf("Starting server at port %s\n", httpPort)
	log.Fatal(http.ListenAndServe(":"+httpPort, nil))
}
