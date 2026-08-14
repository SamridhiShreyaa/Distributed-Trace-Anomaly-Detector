package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Peter/trace-detector/demo/shared"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	shutdown := shared.InitTracer("auth")
	defer shutdown()

	handler := otelhttp.NewHandler(http.HandlerFunc(checkoutHandler), "checkout")
	http.Handle("/checkout", handler)

	log.Println("auth service listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

// checkoutHandler handles /checkout by forwarding the request downstream to the
// cart service, propagating the trace context so the spans are linked.
func checkoutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	client := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   5 * time.Second,
	}

	cartURL := os.Getenv("UPSTREAM_CART_URL")
	if cartURL == "" {
		cartURL = "http://localhost:8082"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", cartURL+"/checkout", nil)
	if err != nil {
		http.Error(w, "failed to build cart request", http.StatusInternalServerError)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "cart service unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	fmt.Fprint(w, "auth: checkout processed\n")
	log.Printf("auth: checkout complete, cart responded %d", resp.StatusCode)
}
