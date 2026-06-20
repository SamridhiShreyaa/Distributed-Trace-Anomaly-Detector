package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("Payment service starting on :8084")
	
	http.HandleFunc("/payment", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "payment_ready"}`))
	})
	
	if err := http.ListenAndServe(":8084", nil); err != nil {
		log.Fatal(err)
	}
}
