// internal/grpc_server.go
package internal

import (
	"database/sql"

	"google.golang.org/grpc"
)

// YourService is an example gRPC service.
type YourService struct {
	// Add your service-specific fields here
}

// Implement your gRPC service methods here

// NewGRPCServer creates and returns a new gRPC server.
func NewGRPCServer(db *sql.DB) *grpc.Server {
	// Set up your gRPC server with the necessary options
	server := grpc.NewServer()

	// Register your gRPC service implementation
	yourService := &YourService{}
	RegisterYourServiceServer(server, yourService)

	return server
}
