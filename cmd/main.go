package main

import (
	"log"
	"net/http"
)

func main() {
	const port = "8080"

	mux := http.NewServeMux()

	// mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	greetings, err := os.ReadFile("assets/index.html")
	// 	if err != nil {
	// 		log.Println("could not find file to serve")
	// 	}
	// 	w.Write(greetings)
	// })

	statDir := http.Dir("./static")
	mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(statDir)))

	assetsDir := http.Dir("./assets")
	mux.Handle("/app/assets/", http.StripPrefix("/app/assets/", http.FileServer(assetsDir)))

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(200)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("OK , We are Chirping"))
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
