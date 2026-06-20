package main

import (
	"flag"
	"log"
	"time"
)

func main() {
	service := flag.String("service", "cart", "Service to target")
	latency := flag.Duration("latency", 300*time.Millisecond, "Latency to inject")
	duration := flag.Duration("duration", 2*time.Minute, "Duration of regression")
	flag.Parse()
	
	log.Printf("Injecting %v latency into %s for %v", latency, *service, *duration)
	
	// TODO: Implement regression injection
	// This will:
	// - Connect to the target service
	// - Inject artificial latency
	// - Run for specified duration
	// - Report back to detector
	
	time.Sleep(*duration)
	log.Println("Regression injection complete")
}
