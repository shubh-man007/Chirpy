package main

import (
	"log"

	"github.com/shubh-man007/Chirpy/cmd/internal/database"
	"github.com/shubh-man007/Chirpy/cmd/internal/server"
)

func main() {
	const port = "8080"

	pgx, err := database.NewDbPgx()
	if err != nil {
		log.Fatalf("failed connecting to DB: %v", err)
	}
	defer pgx.Close()

	log.Print("connected to DB")

	srv := server.New(port, pgx.Queries)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
