package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"

	_ "github.com/lib/pq"
)

// Define CreateNewUserParams struct
type CreateNewUserParams struct {
	ClientID       string
	RateLimit      int32
	RefillInterval int32
}

var testQueries *Queries

func TestMain(m *testing.M) {
	pgConnStr := "postgresql://root:secret@localhost:5432/ratelimitingdb?sslmode=disable"

	conn, err := sql.Open("pgx", pgConnStr)
	if err != nil {
		log.Fatalf("Veritabanına bağlanılamadı: %v", err)
	}
	testQueries = New(conn)
	defer conn.Close()

	os.Exit(m.Run())
}

func TestDatabaseFunctions(t *testing.T) {
	// You can create test data for your tests.
	createParams := CreateRateLimitParams{
		Clientid:       "testClient35",
		RateLimit:      100,
		RefillInterval: 60,
	}

	// Test CreateRateLimit
	createdRateLimit, err := testQueries.CreateRateLimit(context.Background(), createParams)
	if err != nil {
		t.Fatalf("Error creating rate limit: %v", err)
	}

	// Test GetRateLimit
	fetchedRateLimit, err := testQueries.GetRateLimit(context.Background(), createdRateLimit.Clientid)
	if err != nil {
		t.Fatalf("Error fetching rate limit: %v", err)
	}

	// Assert that the fetched rate limit matches the created one.
	if fetchedRateLimit.Clientid != createParams.Clientid ||
		fetchedRateLimit.RateLimit != createParams.RateLimit ||
		fetchedRateLimit.RefillInterval != createParams.RefillInterval {
		t.Errorf("Fetched rate limit does not match the created one.")
	}

}
func TestDeleteID(t *testing.T) {
	deleted_Clientid := "testClient23"
	err := testQueries.DeleteRateLimit(context.Background(), deleted_Clientid)
	if err != nil {
		t.Errorf("Error deleting rate limit: %v", err)
	}

}

func TestListRateLimits(t *testing.T) {

	// Örnek sorgu parametreleri
	params := ListRateLimitsParams{
		Limit:  100,
		Offset: 0,
	}
	result, err := testQueries.ListRateLimits(context.Background(), params)
	if err != nil {
		t.Errorf("ListRateLimits sorgusu hata verdi: %v", err)
	} else {
		// resultı ekrana çıkar
		t.Logf("Result: %+v\n", result)
	}
}
func TestDeleteAll(t *testing.T) {

	err := testQueries.DeleteAllRateLimits(context.Background())
	if err != nil {
		t.Fatalf("Error deleting all rate limits: %v", err)
	}

}
func TestDatabaseFunctionsMultiInput(t *testing.T) {
	// Rastgele sayılar üretebilmek için rastgele bir seed ayarlayın.
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 10; i++ {
		// Rastgele RateLimit ve RefillInterval değerleri oluşturun.
		rateLimit := int32(rand.Intn(21))           // 0-20 arası rastgele bir değer
		refillInterval := int32(rand.Intn(51) + 10) // 10-60 arası rastgele bir değer

		createParams := CreateRateLimitParams{
			Clientid:       fmt.Sprintf("testClient%d", i+1),
			RateLimit:      rateLimit,
			RefillInterval: refillInterval,
		}

		// Test CreateRateLimit
		createdRateLimit, err := testQueries.CreateRateLimit(context.Background(), createParams)
		if err != nil {
			t.Fatalf("Error creating rate limit: %v", err)
		}

		// Test GetRateLimit
		fetchedRateLimit, err := testQueries.GetRateLimit(context.Background(), createdRateLimit.Clientid)
		if err != nil {
			t.Fatalf("Error fetching rate limit: %v", err)
		}

		// Assert that the fetched rate limit matches the created one.
		if fetchedRateLimit.Clientid != createParams.Clientid ||
			fetchedRateLimit.RateLimit != rateLimit ||
			fetchedRateLimit.RefillInterval != refillInterval {
			t.Errorf("Fetched rate limit does not match the created one.")
		}
	}
}
func TestUpdateRateLimit(t *testing.T) {

	// Convert to UpdateRateLimitParams with ClientID included
	toUpdateParams := UpdateRateLimitParams{
		Clientid:       "testClient12", // Correct field name to 'Clientid'
		RateLimit:      100,
		RefillInterval: 13,
	}

	CurrentParameters, err := testQueries.GetRateLimit(context.Background(), toUpdateParams.Clientid)
	if err != nil {
		t.Fatalf("GetRateLimit query failed: %v", err)
	}

	// Call UpdateRateLimit function
	err = testQueries.UpdateRateLimit(context.Background(), toUpdateParams)

	// Error check
	if err != nil {
		t.Fatalf("UpdateRateLimit query failed: %v", err)
	}

	// Get updated data from the database
	UpdatedParams, err := testQueries.GetRateLimit(context.Background(), toUpdateParams.Clientid)

	// Error check
	if err != nil {
		t.Fatalf("GetRateLimit query failed: %v", err)
	}

	// Check updated data
	if CurrentParameters.RateLimit == UpdatedParams.RateLimit ||
		CurrentParameters.RefillInterval == UpdatedParams.RefillInterval {
		t.Errorf("Cannot Update Parameters!, they are the same")
	}
}
