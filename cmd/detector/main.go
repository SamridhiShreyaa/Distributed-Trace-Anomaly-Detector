package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v", sig)
		cancel()
	}()

	log.Println("Distributed Trace Anomaly Detector starting...")

	// TODO: Initialize components
	// - Config loading
	// - Storage initialization
	// - OTLP receiver setup
	// - Baseline model initialization
	// - Detector startup
	// - API server startup

	fmt.Println("Detector initialized successfully")

	<-ctx.Done()
	log.Println("Detector shutting down...")
}
