package middleware

import (
	"log"
	"net/http"

	"github.com/shubh-man007/Chirpy/cmd/internal/config"
)

func HitCounterMiddleware(cfg *config.ApiConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			cfg.FileserverHits.Add(1)
			log.Print("[COUNT] app requested")
		}

		next.ServeHTTP(w, r)
	})
}
