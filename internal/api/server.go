package api

import (
	"context"
)

// Server provides gRPC and REST APIs for querying detector results
type Server struct {
	address string
	port    int
}

// NewServer creates a new API server
func NewServer(address string, port int) *Server {
	return &Server{
		address: address,
		port:    port,
	}
}

// Start begins serving API requests
func (s *Server) Start(ctx context.Context) error {
	// TODO: Implement gRPC server
	// TODO: Implement REST gateway with grpc-gateway
	// This will expose:
	// - StreamAnomalies - stream recent anomalies
	// - GetTrace - query a specific trace
	// - GetBaseline - inspect baseline statistics
	return nil
}

// Stop gracefully shuts down the API server
func (s *Server) Stop(ctx context.Context) error {
	// TODO: Implement graceful shutdown
	return nil
}
