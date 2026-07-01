package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Peter/trace-detector/demo/shared"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	shutdown := shared.InitTracer("payment")
	defer shutdown()

	handler := otelhttp.NewHandler(http.HandlerFunc(checkoutHandler), "checkout")
	http.Handle("/checkout", handler)

	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8084"
	}

	log.Printf("payment service listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// checkoutHandler simulates processing a charge, optionally padded with
// artificial latency (via ARTIFICIAL_LATENCY_MS) to exercise the anomaly
// detector.
func checkoutHandler(w http.ResponseWriter, r *http.Request) {
	// ARTIFICIAL_LATENCY_MS is optional; any unset/invalid value means 0.
	artificialMs, _ := strconv.Atoi(os.Getenv("ARTIFICIAL_LATENCY_MS"))

	// base work — represents a normal charge
	time.Sleep(40 * time.Millisecond)

	if artificialMs > 0 {
		time.Sleep(time.Duration(artificialMs) * time.Millisecond)
	}

	w.Write([]byte("payment: charge processed\n"))
	log.Printf("payment: charge processed, artificial latency %dms", artificialMs)
}
