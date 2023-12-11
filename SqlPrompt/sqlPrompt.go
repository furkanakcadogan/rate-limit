// SqlPrompt/sqlPrompt.go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/furkanakcadogan/rate-limit/db"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// Define CreateNewUserParams struct
type CreateNewUserParams struct {
	ClientID       string
	RateLimit      int32
	RefillInterval int32
}

func InsertNewClient(queries *db.Queries) {
	fmt.Printf("Inserting new Client:\n")
	// Prompt the user for clientid
	fmt.Print("Enter clientid: ")
	var clientid string
	_, err := fmt.Scan(&clientid)
	if err != nil {
		log.Printf("Error reading clientid: %v", err)
		return
	}

	// Check if the clientid already exists in the database
	_, err = queries.GetRateLimit(context.Background(), clientid)
	if err == nil {
		// Clientid already exists, so we skip the duplicate entry
		log.Printf("Clientid %s already exists in the database. Skipping duplicate entry.", clientid)
		return
	} else if err != sql.ErrNoRows {
		// An error occurred other than "not found" error
		log.Printf("Error checking if clientid exists: %v", err)
		return
	}

	// Prompt the user for rate limit and refill interval
	fmt.Print("Enter Rate Limit: ")
	var rateLimit int32
	_, err = fmt.Scanf("%d", &rateLimit)
	if err != nil {
		log.Printf("Error reading Rate Limit: %v", err)
		return
	}

	fmt.Print("Enter Refill Interval: ")
	var refillInterval int32
	_, err = fmt.Scanf("%d", &refillInterval)
	if err != nil {
		log.Printf("Error reading Refill Interval: %v", err)
		return
	}

	// Create CreateRateLimitParams
	createParams := db.CreateRateLimitParams{
		Clientid:       clientid,
		RateLimit:      rateLimit,
		RefillInterval: refillInterval,
	}

	// Test CreateRateLimit
	createdRateLimit, err := queries.CreateRateLimit(context.Background(), createParams)
	if err != nil {
		log.Printf("Error creating rate limit: %v", err)
		return
	}

	// Test GetRateLimit
	fetchedRateLimit, err := queries.GetRateLimit(context.Background(), createdRateLimit.Clientid)
	if err != nil {
		log.Printf("Error fetching rate limit: %v", err)
		return
	}

	// Assert that the fetched rate limit matches the created one.
	if fetchedRateLimit.Clientid != createParams.Clientid ||
		fetchedRateLimit.RateLimit != createParams.RateLimit ||
		fetchedRateLimit.RefillInterval != createParams.RefillInterval {
		log.Printf("Fetched rate limit does not match the created one.")
	}
}

// Helper function to check if an error is a duplicate key violation
func isDuplicateKeyError(err error) bool {
	if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
		return true
	}
	return false
}

func DeleteClient(queries *db.Queries) {
	fmt.Printf("Deleting Existing Client:\n")
	deleted_Clientid := "testClient23"
	err := queries.DeleteRateLimit(context.Background(), deleted_Clientid)
	if err != nil {
		log.Fatalf("Error deleting rate limit: %v", err)
	}
}
func TestMultiplyAllRateLimits(queries *db.Queries) {

}
func ListClientIDRecords(queries *db.Queries) {
	fmt.Printf("Listing Existing Records:\n")
	// Sample query parameters
	params := db.ListRateLimitsParams{
		Limit:  100,
		Offset: 0,
	}

	result, err := queries.ListRateLimits(context.Background(), params)
	if err != nil {
		log.Fatalf("ListRateLimits query failed: %v", err)
	} else {
		// Print each result on a new line
		for _, rateLimit := range result {
			log.Printf("ID: %d, ClientID: %s, RateLimit: %d, RefillInterval: %d",
				rateLimit.ID, rateLimit.Clientid, rateLimit.RateLimit, rateLimit.RefillInterval)
		}
	}
}

func GenerateRandomClientIDs(queries *db.Queries) {
	fmt.Printf("Generate Random ClientIDs Prompt: (Warning! Due to nature of random values you can view some duplication errors)\n")
	// Prompt the user for the number of records to create
	fmt.Print("Enter the number of records to create: ")
	var numRecords int
	_, err := fmt.Scanf("%d", &numRecords)
	if err != nil {
		log.Printf("Error reading input: %v", err)
		return // Burada return kullanarak hatalı giriş durumunda fonksiyonu sonlandırabilirsiniz.
	}

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < numRecords; i++ {
		rateLimit := int32(rand.Intn(21))                   // 0-20 arası rastgele bir değer
		refillInterval := int32(rand.Intn(51) + 10)         // 10-60 arası rastgele bir değer
		clientID := fmt.Sprintf("Client%d", rand.Intn(101)) // 0 ile 100 arasında rastgele bir değer al

		createParams := db.CreateRateLimitParams{
			Clientid:       clientID,
			RateLimit:      rateLimit,
			RefillInterval: refillInterval,
		}

		createdRateLimit, err := queries.CreateRateLimit(context.Background(), createParams)
		if err != nil {
			if isDuplicateKeyError(err) {
				log.Printf("Skipping record creation for %s due to duplicate key violation.", createParams.Clientid)
				continue // Skip this iteration and proceed to the next record
			}
			log.Printf("Error creating rate limit: %v", err)
			continue // Burada da continue kullanarak hata durumunda sonraki kayda geçebilirsiniz.
		}

		fetchedRateLimit, err := queries.GetRateLimit(context.Background(), createdRateLimit.Clientid)
		if err != nil {
			log.Printf("Error fetching rate limit: %v", err)
			continue // Hata durumunda sonraki kayda geç
		}

		if fetchedRateLimit.Clientid != createParams.Clientid ||
			fetchedRateLimit.RateLimit != rateLimit ||
			fetchedRateLimit.RefillInterval != refillInterval {
			log.Printf("Fetched rate limit does not match the created one.")
			continue // Veri uyumsuzluğu durumunda sonraki kayda geç
		}
	}
}

func UpdateClientParameters(queries *db.Queries) {
	fmt.Printf("Updating Client Parameters:\n")
	// Prompt the user for clientid
	fmt.Print("Enter clientid to update: ")
	var clientid string
	_, err := fmt.Scan(&clientid)
	if err != nil {
		log.Printf("Error reading clientid: %v", err)
		return
	}

	// Get current rate limit and refill interval for the specified clientid
	currentParams, err := queries.GetRateLimit(context.Background(), clientid)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("Client %s not found in the database. Continuing...\n", clientid)
			return // Client not found, so we don't prompt for new values
		} else {
			log.Printf("GetRateLimit query failed: %v", err)
			return
		}
	}

	// Client is found, so we prompt for new values
	fmt.Printf("Current Rate Limit for %s: %d\n", clientid, currentParams.RateLimit)
	fmt.Printf("Current Refill Interval for %s: %d\n", clientid, currentParams.RefillInterval)

	// Prompt the user for new rate limit and refill interval
	fmt.Print("Enter new Rate Limit: ")
	var newRateLimit int32
	_, err = fmt.Scanf("%d", &newRateLimit)
	if err != nil {
		log.Printf("Error reading new Rate Limit: %v", err)
		return
	}

	fmt.Print("Enter new Refill Interval: ")
	var newRefillInterval int32
	_, err = fmt.Scanf("%d", &newRefillInterval)
	if err != nil {
		log.Printf("Error reading new Refill Interval: %v", err)
		return
	}

	// Create UpdateRateLimitParams
	updateParams := db.UpdateRateLimitParams{
		Clientid:       clientid,
		RateLimit:      newRateLimit,
		RefillInterval: newRefillInterval,
	}

	// Call UpdateRateLimit function
	err = queries.UpdateRateLimit(context.Background(), updateParams)

	// Error check
	if err != nil {
		log.Printf("UpdateRateLimit query failed: %v", err)
		return
	}

	// Get updated data from the database
	updatedParams, err := queries.GetRateLimit(context.Background(), clientid)

	// Error check
	if err != nil {
		log.Printf("GetRateLimit query failed: %v", err)
		return
	}

	// Print updated data
	fmt.Printf("Updated Rate Limit for %s: %d\n", clientid, updatedParams.RateLimit)
	fmt.Printf("Updated Refill Interval for %s: %d\n", clientid, updatedParams.RefillInterval)
}

func askForConfirmation() bool {
	fmt.Print("Are you sure? (y/n): ")
	var response string
	_, err := fmt.Scan(&response)
	if err != nil {
		log.Printf("Error reading input: %v", err)
		return false
	}
	return strings.ToLower(response) == "y"
}

func DeleteAllClients(queries *db.Queries) {
	fmt.Printf("Deleting all database:\n")
	if !askForConfirmation() {
		fmt.Println("Operation canceled.")
		return
	}

	err := queries.DeleteAllRateLimits(context.Background())
	if err != nil {
		log.Fatalf("Error deleting all rate limits: %v", err)
	}
	fmt.Println("All rate limits deleted successfully.")
}

func main() {
	pgConnStr := "postgresql://root:secret@localhost:5432/ratelimitingdb?sslmode=disable"
	conn, err := sql.Open("pgx", pgConnStr)
	if err != nil {
		log.Fatalf("Veritabanına bağlanılamadı: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)

	for {
		fmt.Println("\nSelect an option:")
		fmt.Println("1. Insert New Client")
		fmt.Println("2. Delete Client")
		fmt.Println("3. List Client ID Records")
		fmt.Println("4. Generate Random Client IDs")
		fmt.Println("5. Update Client Parameters")
		fmt.Println("6. Delete All Clients")
		fmt.Println("7. Quit")

		var choice int
		fmt.Print("Enter your choice (1-8): ")
		_, err := fmt.Scanf("%d", &choice)
		if err != nil {
			log.Printf("Error reading choice: %v", err)
			continue
		}

		switch choice {
		case 1:
			InsertNewClient(queries)
		case 2:
			DeleteClient(queries)
		case 3:
			ListClientIDRecords(queries)
		case 4:
			GenerateRandomClientIDs(queries)
		case 5:
			UpdateClientParameters(queries)
		case 6:
			DeleteAllClients(queries)
		case 7:
			fmt.Println("Exiting the program.")
			return
		default:
			fmt.Println("Invalid choice. Please select a valid option (1-7).")
		}
	}

}
