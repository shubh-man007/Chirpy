package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
)

type apiConfig struct {
	rootHits atomic.Int32
}

func (cfg *apiConfig) HitCounterMiddleware(next http.Handler) http.Handler {
	newHit := cfg.rootHits.Add(1)
	cfg.rootHits.Store(newHit)
	return next
}

func main() {
	const port = "8080"

	cfg := &apiConfig{}
	mux := http.NewServeMux()

	statDir := http.Dir("./static")
	mux.Handle("/app/", http.StripPrefix("/app/", cfg.HitCounterMiddleware(http.FileServer(statDir))))

	assetsDir := http.Dir("./assets")
	mux.Handle("/app/assets/", http.StripPrefix("/app/assets/", http.FileServer(assetsDir)))

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("OK , We are Chirping"))
		}
	})

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			hits := cfg.rootHits.Load()
			strHits := fmt.Sprintf("Hits: %s", strconv.Itoa(int(hits)))
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(strHits))
		}
	})

	mux.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			cfg.rootHits.Store(0)
			hits := cfg.rootHits.Load()
			strHits := fmt.Sprintf("Hits: %s", strconv.Itoa(int(hits)))
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(strHits))
		}
	})

	s := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files on port: %s\n", port)

	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("Failed to listen at port %s. Error: %v", port, err)
	}
}
