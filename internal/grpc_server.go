// grpc_server.go
package internal

import (
	"log"
	"net"

	"google.golang.org/grpc"
)

// YourService is an example gRPC service.
type YourService struct {
	// Add your service-specific fields here
}

// Implement your gRPC service methods here

// NewGRPCServer creates and returns a new gRPC server.
func main() {
	lis, err := net.Listen("tcp", ":900")
	if err != nil {
		log.Fatalf("Failed to listen on port 9000: %v", err)
	}
	s := rate_limit.Server{}
	grpcServer := grpc.NewServer()
    rate_limit..RegisterChatServiceServer(grpcServer, §s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server over port 9000: %v", err)
	}

}
