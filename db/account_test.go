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
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	testQueries = New(conn)
	defer conn.Close()

	os.Exit(m.Run())
}

func TestDatabaseFunctions(t *testing.T) {
	createParams := CreateRateLimitParams{
		Clientid:       "testClient35",
		RateLimit:      100,
		RefillInterval: 60,
	}

	createdRateLimit, err := testQueries.CreateRateLimit(context.Background(), createParams)
	if err != nil {
		t.Fatalf("Error creating rate limit: %v", err)
	}

	fetchedRateLimit, err := testQueries.GetRateLimit(context.Background(), createdRateLimit.Clientid)
	if err != nil {
		t.Fatalf("Error fetching rate limit: %v", err)
	}

	if fetchedRateLimit.Clientid != createParams.Clientid ||
		fetchedRateLimit.RateLimit != createParams.RateLimit ||
		fetchedRateLimit.RefillInterval != createParams.RefillInterval {
		t.Errorf("Fetched rate limit does not match the created one.")
	}
}

func TestDeleteID(t *testing.T) {
	deletedClientID := "testClient23"
	err := testQueries.DeleteRateLimit(context.Background(), deletedClientID)
	if err != nil {
		t.Errorf("Error deleting rate limit: %v", err)
	}
}

func TestListRateLimits(t *testing.T) {
	params := ListRateLimitsParams{
		Limit:  100,
		Offset: 0,
	}
	result, err := testQueries.ListRateLimits(context.Background(), params)
	if err != nil {
		t.Errorf("ListRateLimits query failed: %v", err)
	} else {
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
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 10; i++ {
		rateLimit := int32(rand.Intn(21))
		refillInterval := int32(rand.Intn(51) + 10)

		createParams := CreateRateLimitParams{
			Clientid:       fmt.Sprintf("testClient%d", i+1),
			RateLimit:      rateLimit,
			RefillInterval: refillInterval,
		}

		createdRateLimit, err := testQueries.CreateRateLimit(context.Background(), createParams)
		if err != nil {
			t.Fatalf("Error creating rate limit: %v", err)
		}

		fetchedRateLimit, err := testQueries.GetRateLimit(context.Background(), createdRateLimit.Clientid)
		if err != nil {
			t.Fatalf("Error fetching rate limit: %v", err)
		}

		if fetchedRateLimit.Clientid != createParams.Clientid ||
			fetchedRateLimit.RateLimit != rateLimit ||
			fetchedRateLimit.RefillInterval != refillInterval {
			t.Errorf("Fetched rate limit does not match the created one.")
		}
	}
}

func TestUpdateRateLimit(t *testing.T) {
	toUpdateParams := UpdateRateLimitParams{
		Clientid:       "testClient12",
		RateLimit:      100,
		RefillInterval: 13,
	}

	CurrentParameters, err := testQueries.GetRateLimit(context.Background(), toUpdateParams.Clientid)
	if err != nil {
		t.Fatalf("GetRateLimit query failed: %v", err)
	}

	err = testQueries.UpdateRateLimit(context.Background(), toUpdateParams)

	if err != nil {
		t.Fatalf("UpdateRateLimit query failed: %v", err)
	}

	UpdatedParams, err := testQueries.GetRateLimit(context.Background(), toUpdateParams.Clientid)

	if err != nil {
		t.Fatalf("GetRateLimit query failed: %v", err)
	}

	if CurrentParameters.RateLimit == UpdatedParams.RateLimit ||
		CurrentParameters.RefillInterval == UpdatedParams.RefillInterval {
		t.Errorf("Cannot Update Parameters!, they are the same")
	}
}

func TestMultiplyAllRateLimits(t *testing.T) {
	arg := MultiplyAllRateLimitsParams{
		RateLimit:      5,
		RefillInterval: 4,
	}

	err := testQueries.MultiplyAllRateLimits(context.Background(), arg)
	if err != nil {
		t.Fatalf("MultiplyAllRateLimits failed: %v", err)
	}

	// You can write additional queries here to validate the changes made in the database, if needed.
}

func TestSafelyMultiplyAllRateLimits(t *testing.T) {
	// Define multipliers here if needed.
}
