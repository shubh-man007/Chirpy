package middleware

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	FileserverHits atomic.Int32
}

func NewApiCfg() *apiConfig {
	return &apiConfig{}
}

func (cfg *apiConfig) HitCounterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			cfg.FileserverHits.Add(1)
			log.Print("COUNT app requested")
		}

		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.URL.Path == "/metrics" {
		x := cfg.FileserverHits.Load()
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(fmt.Sprintf("Hits: %d", x)))
		if err != nil {
			log.Printf("could not write to metric endpoint: %s", err)
		}
		return
	}

	if r.Method == http.MethodPost && r.URL.Path == "/reset" {
		cfg.FileserverHits.Store(0)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Hits: 0"))
		if err != nil {
			log.Printf("could not write to reset endpoint: %s", err)
		}
		return
	}

	http.NotFound(w, r)
}
