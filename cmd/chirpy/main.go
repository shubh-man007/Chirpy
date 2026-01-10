package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/shubh-man007/Chirpy/cmd/internal/database"
	"github.com/shubh-man007/Chirpy/cmd/internal/server"
)

func main() {
	const port = "8080"

	err := godotenv.Load()
	if err != nil {
		log.Printf("could not load .env file: %s", err)
	}

	connStr := os.Getenv("DB_URL")

	pgx, err := database.NewDbPgx(connStr)
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
