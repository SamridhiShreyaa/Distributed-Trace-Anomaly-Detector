package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("Inventory service starting on :8083")
	
	http.HandleFunc("/inventory", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "inventory_ready"}`))
	})
	
	if err := http.ListenAndServe(":8083", nil); err != nil {
		log.Fatal(err)
	}
}
