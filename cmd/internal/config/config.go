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
}

func NewApiCfg(db *database.Queries, platform string, jwt_secret string) *ApiConfig {
	return &ApiConfig{
		DB:        db,
		Platform:  platform,
		JWTSecret: jwt_secret,
	}
}
