package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("Auth service starting on :8081")
	
	http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "authenticated"}`))
	})
	
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}
}
