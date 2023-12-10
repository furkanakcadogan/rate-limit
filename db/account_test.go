package db

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"

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
	pgConnStr := "postgresql://root:secret@localhost:5432/rate_limiting_db?sslmode=disable"

	conn, err := sql.Open("postgres", pgConnStr)
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
		Clientid:       "testClient",
		RateLimit:      100,
		RefillInterval: 60,
	}

	// Test CreateRateLimit
	createdRateLimit, err := testQueries.CreateRateLimit(context.Background(), createParams)
	if err != nil {
		t.Fatalf("Error creating rate limit: %v", err)
	}

	// Test GetRateLimit
	fetchedRateLimit, err := testQueries.GetRateLimit(context.Background(), createdRateLimit.ID)
	if err != nil {
		t.Fatalf("Error fetching rate limit: %v", err)
	}

	// Assert that the fetched rate limit matches the created one.
	if fetchedRateLimit.Clientid != createParams.Clientid ||
		fetchedRateLimit.RateLimit != createParams.RateLimit ||
		fetchedRateLimit.RefillInterval != createParams.RefillInterval {
		t.Errorf("Fetched rate limit does not match the created one.")
	}

	// You can continue with similar tests for other functions like ListRateLimits and UpdateRateLimit.

	// Clean up by deleting the rate limit created for testing.
	err = testQueries.DeleteRateLimit(context.Background(), createdRateLimit.ID)
	if err != nil {
		t.Errorf("Error deleting rate limit: %v", err)
	}
}
