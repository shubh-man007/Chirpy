package main

import (
	"log"

	"github.com/shubh-man007/Chirpy/cmd/internal/server"
)

func main() {
	const port = "8080"

	srv := server.New(port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
