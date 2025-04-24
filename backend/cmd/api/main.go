package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Basic HTTP server setup (replace with your actual API logic)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	port := "8080" // Example port, consider making this configurable
	fmt.Printf("Backend server starting on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("could not start server: %v\n", err)
	}
} 