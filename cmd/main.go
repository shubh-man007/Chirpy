package main

import (
	"log"
	"net/http"

	"github.com/shubh-man007/Chirpy/cmd/internal/middleware"
)

const readinessMessage = "OK"

func main() {
	const port = "8080"
	mux := http.NewServeMux()

	cfg := middleware.NewApiCfg()
	cfg.FileserverHits.Store(0)

	mux.Handle("/app/", http.StripPrefix("/app/", cfg.HitCounterMiddleware(http.FileServer(http.Dir("./static")))))
	mux.Handle("/app/assets/", http.StripPrefix("/app/assets/", http.FileServer(http.Dir("./assets"))))

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(200)
			w.Header().Add("Content-Type", "text/plain")
			w.Header().Add("Content-Type", "charset=utf-8")
			_, err := w.Write([]byte(readinessMessage))
			if err != nil {
				log.Printf("could not write to readiness endpoint: %s", err.Error())
			}
		}
	})

	mux.HandleFunc("GET /metrics", cfg.ServeHTTP)
	mux.HandleFunc("POST /reset", cfg.ServeHTTP)

	s := &http.Server{
		Addr:    ":" + port,
		Handler: middleware.LogMiddleware(mux),
	}

	log.Printf("Running server at port:%s", port)

	err := s.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to listen at port %s. Error: %v", port, err)
	}
}
