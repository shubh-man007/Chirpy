package middleware

import (
	"log"
	"net/http"
	"sync/atomic"
)

type ApiConfig struct {
	FileserverHits atomic.Int32
}

func NewApiCfg() *ApiConfig {
	return &ApiConfig{}
}

func (cfg *ApiConfig) HitCounterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			cfg.FileserverHits.Add(1)
			log.Print("[COUNT] app requested")
		}

		next.ServeHTTP(w, r)
	})
}
