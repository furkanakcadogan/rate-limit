package grpc

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/furkanakcadogan/rate-limit/proto"
)

// RateLimitServer struct implements the generated gRPC server interface
type RateLimitServer struct {
	proto.UnimplementedRateLimitServiceServer // Embed the Unimplemented server
}

// Implement the CheckRateLimit RPC method
func (s *RateLimitServer) CheckRateLimit(ctx context.Context, req *proto.RateLimitRequest) (*proto.RateLimitResponse, error) {
	log.Printf("Received RateLimitRequest: %+v\n", req)

	// Implement your rate-limiting logic here

	// For demonstration purposes, always allow the request and return a response
	response := &proto.RateLimitResponse{
		Allowed:         true,
		RemainingTokens: req.TokensRequired,
	}

	return response, nil
}

func grpc_server() {
	// Create a new gRPC server
	grpcServer := grpc.NewServer()

	// Register your service implementation
	proto.RegisterRateLimitServiceServer(grpcServer, &RateLimitServer{})

	// Start the server on a specific port
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Println("Server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
