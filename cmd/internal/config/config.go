package config

import (
	"sync/atomic"

	"github.com/shubh-man007/Chirpy/cmd/internal/database"
)

type ApiConfig struct {
	FileserverHits atomic.Int32
	DB             *database.Queries
	Platform       string
	JWTSecret      string
	PolkaAPIKey    string
}

func NewApiCfg(db *database.Queries, platform string, jwtSecret string, polkaAPI string) *ApiConfig {
	return &ApiConfig{
		DB:          db,
		Platform:    platform,
		JWTSecret:   jwtSecret,
		PolkaAPIKey: polkaAPI,
	}
}
