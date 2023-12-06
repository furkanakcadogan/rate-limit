// cmd/main.go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	// Import other necessary packages
)

func main() {
	// Connect to PostgreSQL database
	db, err := sql.Open("postgres", "user=rate_limit_user dbname=rate_limit_db sslmode=disable password=your_password")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize your gRPC server with rate limiting middleware
	server := internal.NewGRPCServer(db)

	// Set up signal handling to gracefully shut down the server on interrupt signals
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start the gRPC server in a separate goroutine
	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	fmt.Println("Server is ready to receive requests.")
	<-interrupt // Wait for interrupt signal

	fmt.Println("Shutting down gracefully...")
	server.GracefulStop()
}
