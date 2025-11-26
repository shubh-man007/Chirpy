package main

import (
	"log"
	"net/http"
)

func main() {
	const port = ":8080"

	mux := http.NewServeMux()

	// mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	greetings, err := os.ReadFile("assets/index.html")
	// 	if err != nil {
	// 		log.Println("could not find file to serve")
	// 	}

	// 	w.Write(greetings)
	// })

	dir := http.Dir("./assets")
	mux.Handle("/", http.FileServer(dir))

	s := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	log.Printf("Listening at port %s", port)

	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("Failed to listen at port %s. Error: %v", port, err)
	}
}
