package main

import (
	"log"
	"net/http"
)

func main() {
	const port = ":8080"

	mux := http.NewServeMux()

	s := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	log.Printf("Listening at port %s", port)

	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("Failed to listen at port %s. Error: %v", port, err)
	}
}
