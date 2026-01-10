package database

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type DbPgx struct {
	DB      *sql.DB
	Queries *Queries
}

func NewDbPgx(connStr string) (*DbPgx, error) {
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
