package main

import (
	"log"

	"github.com/shubh-man007/Chirpy/cmd/internal/server"
)

func main() {
	const port = "8080"

	// TODO: Initialize database connection when ready
	// db, err := database.OpenDB(...)
	// if err != nil {
	// 	log.Fatalf("Failed to open database: %v", err)
	// }
	// defer db.Close()
	// queries := database.New(db)

	srv := server.New(port, nil)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
