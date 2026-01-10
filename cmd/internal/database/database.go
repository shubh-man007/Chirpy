package database

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type DbPgx struct {
	DB      *sql.DB
	Queries *Queries
}

func NewDbPgx() (*DbPgx, error) {
	err := godotenv.Load()
	if err != nil {
		log.Printf("could not load .env file: %s", err)
		return nil, err
	}

	connStr := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	queries := New(db)

	return &DbPgx{
		DB:      db,
		Queries: queries,
	}, nil
}

func (d *DbPgx) Close() error {
	return d.DB.Close()
}
