package config

import (
	"sync/atomic"

	"github.com/shubh-man007/Chirpy/cmd/internal/database"
)

type ApiConfig struct {
	FileserverHits atomic.Int32
	DB             *database.Queries
}

func NewApiCfg(db *database.Queries) *ApiConfig {
	return &ApiConfig{
		DB: db,
	}
}

