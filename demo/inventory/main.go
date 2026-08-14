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
	shutdown := shared.InitTracer("inventory")
	defer shutdown()

	handler := otelhttp.NewHandler(http.HandlerFunc(checkoutHandler), "checkout")
	http.Handle("/checkout", handler)

	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8083"
	}

	log.Printf("inventory service listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// checkoutHandler simulates an inventory lookup, optionally padded with
// artificial latency (via ARTIFICIAL_LATENCY_MS) to exercise the anomaly
// detector.
func checkoutHandler(w http.ResponseWriter, r *http.Request) {
	// ARTIFICIAL_LATENCY_MS is optional; any unset/invalid value means 0.
	artificialMs, _ := strconv.Atoi(os.Getenv("ARTIFICIAL_LATENCY_MS"))

	// base work — represents a normal inventory lookup
	time.Sleep(50 * time.Millisecond)

	if artificialMs > 0 {
		time.Sleep(time.Duration(artificialMs) * time.Millisecond)
	}

	w.Write([]byte("inventory: stock checked\n"))
	log.Printf("inventory: stock checked, artificial latency %dms", artificialMs)
}
