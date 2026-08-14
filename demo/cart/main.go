package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/Peter/trace-detector/demo/shared"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	shutdown := shared.InitTracer("cart")
	defer shutdown()

	handler := otelhttp.NewHandler(http.HandlerFunc(checkoutHandler), "checkout")
	http.Handle("/checkout", handler)

	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8082"
	}

	log.Printf("cart service listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// checkoutHandler fans out to the inventory and payment services in parallel,
// propagating the trace context to both, and fails the checkout if either is
// unavailable.
func checkoutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	inventoryURL := os.Getenv("UPSTREAM_INVENTORY_URL")
	if inventoryURL == "" {
		inventoryURL = "http://localhost:8083"
	}
	paymentURL := os.Getenv("UPSTREAM_PAYMENT_URL")
	if paymentURL == "" {
		paymentURL = "http://localhost:8084"
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Each goroutine writes only to its own error variable, so no
	// synchronization beyond the WaitGroup is required.
	var inventoryErr, paymentErr error
	go func() {
		defer wg.Done()
		inventoryErr = callService(ctx, inventoryURL+"/checkout")
	}()
	go func() {
		defer wg.Done()
		paymentErr = callService(ctx, paymentURL+"/checkout")
	}()

	wg.Wait()

	if inventoryErr != nil {
		http.Error(w, "inventory service unavailable", http.StatusBadGateway)
		return
	}
	if paymentErr != nil {
		http.Error(w, "payment service unavailable", http.StatusBadGateway)
		return
	}

	fmt.Fprint(w, "cart: checkout processed\n")
	log.Println("cart: checkout complete, inventory and payment ok")
}

// callService issues an instrumented GET to url, propagating the trace context.
// It returns an error if the request cannot be built or the call fails.
func callService(ctx context.Context, url string) error {
	client := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
