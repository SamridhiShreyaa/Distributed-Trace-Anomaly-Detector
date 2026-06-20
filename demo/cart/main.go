package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("Cart service starting on :8082")
	
	http.HandleFunc("/cart", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "cart_ready"}`))
	})
	
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatal(err)
	}
}
