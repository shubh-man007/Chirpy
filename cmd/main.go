// package main

// import (
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"strconv"
// 	"sync/atomic"
// )

// type apiConfig struct {
// 	rootHits atomic.Int32
// }

// func (cfg *apiConfig) HitCounterMiddleware(next http.Handler) http.Handler {
// 	newHit := cfg.rootHits.Add(1)
// 	cfg.rootHits.Store(newHit)
// 	return next
// }

// func main() {
// 	const port = "8080"

// 	cfg := &apiConfig{}
// 	mux := http.NewServeMux()

// 	statDir := http.Dir("./static")
// 	mux.Handle("/app/", http.StripPrefix("/app/", cfg.HitCounterMiddleware(http.FileServer(statDir))))

// 	assetsDir := http.Dir("./assets")
// 	mux.Handle("/app/assets/", http.StripPrefix("/app/assets/", http.FileServer(assetsDir)))

// 	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method == http.MethodGet {
// 			w.WriteHeader(http.StatusOK)
// 			w.Header().Set("Content-Type", "text/plain")
// 			w.Write([]byte("OK , We are Chirping"))
// 		}
// 	})

// 	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method == http.MethodGet {
// 			hits := cfg.rootHits.Load()
// 			strHits := fmt.Sprintf("Hits: %s", strconv.Itoa(int(hits)))
// 			w.WriteHeader(http.StatusOK)
// 			w.Header().Set("Content-Type", "text/plain")
// 			w.Write([]byte(strHits))
// 		}
// 	})

// 	mux.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method == http.MethodGet {
// 			cfg.rootHits.Store(0)
// 			hits := cfg.rootHits.Load()
// 			strHits := fmt.Sprintf("Hits: %s", strconv.Itoa(int(hits)))
// 			w.WriteHeader(http.StatusOK)
// 			w.Header().Set("Content-Type", "text/plain")
// 			w.Write([]byte(strHits))
// 		}
// 	})

// 	s := &http.Server{
// 		Addr:    ":" + port,
// 		Handler: mux,
// 	}

// 	log.Printf("Serving files on port: %s\n", port)

// 	if err := s.ListenAndServe(); err != nil {
// 		log.Fatalf("Failed to listen at port %s. Error: %v", port, err)
// 	}
// }

package main

import (
	"log"
	"net/http"

	"github.com/shubh-man007/Chirpy/cmd/internal/middleware"
)

const readinessMessage = `
<html>
<head>
	<meta charset="UTF-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1.0" />
	<link rel="icon" href="./favicon.ico" type="image/x-icon">
	<title>Chirpy</title>
	<link rel="preconnect" href="https://fonts.googleapis.com">
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
	<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600&family=Inter:wght@400;500;600&display=swap" rel="stylesheet">
</head>
<body>
	<h3>OK</h3>
</body>
</html>
`

func main() {
	const port = "8080"
	mux := http.NewServeMux()

	mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("./static"))))
	mux.Handle("/app/assets/", http.StripPrefix("/app/assets/", http.FileServer(http.Dir("./assets"))))

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
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
