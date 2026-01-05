package middleware

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

const metricBody = `
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
`

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
			log.Print("[COUNT] app requested")
		}

		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.URL.Path == "/admin/metrics" {
		x := cfg.FileserverHits.Load()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(fmt.Sprintf(metricBody, x)))
		if err != nil {
			log.Printf("could not write to metric endpoint: %s", err)
		}
		return
	}

	if r.Method == http.MethodPost && r.URL.Path == "/admin/reset" {
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
