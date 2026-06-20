package ingest

import (
	"context"
)

// OTLPReceiver ingests OpenTelemetry Protocol traces
type OTLPReceiver struct {
	address string
	port    int
}

// NewOTLPReceiver creates a new OTLP receiver
func NewOTLPReceiver(address string, port int) *OTLPReceiver {
	return &OTLPReceiver{
		address: address,
		port:    port,
	}
}

// Start begins listening for incoming traces
func (r *OTLPReceiver) Start(ctx context.Context) error {
	// TODO: Implement gRPC server for OTLP traces
	// This will:
	// - Listen on gRPC port
	// - Convert OTLP spans to internal Span type
	// - Feed spans into the TraceCollector
	return nil
}

// Stop gracefully shuts down the receiver
func (r *OTLPReceiver) Stop(ctx context.Context) error {
	// TODO: Implement graceful shutdown
	return nil
}
