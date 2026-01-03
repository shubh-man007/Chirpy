// package main

// import (
// 	"fmt"
// 	"time"
// )

// type Request struct {
// 	Path string
// }

// func handleRequest(req Request) {
// 	fmt.Printf("Processing %s\n", req.Path)
// 	time.Sleep(2 * time.Second)
// 	fmt.Printf("Processed %s\n", req.Path)
// }

// func handleRequests(reqs <-chan Request) {
// 	for req := range reqs {
// 		go handleRequest(req)
// 	}
// }

// func main() {
// 	reqs := make(chan Request, 100)
// 	go handleRequests(reqs)
// 	for i := 0; i < 4; i++ {
// 		reqs <- Request{Path: fmt.Sprintf("/path/%d", i)}
// 		time.Sleep(500 * time.Millisecond)
// 	}

// 	time.Sleep(5 * time.Second)
// 	fmt.Println("5 seconds passed, killing server")
// }

package main

import (
	"log"
	"net/http"
)

func fooHandle(w http.ResponseWriter, r *http.Request) {
	log.Print("Executing foo endpoint")
	w.Write([]byte("OK"))
}

func barHandle(w http.ResponseWriter, r *http.Request) {
	log.Print("Executing bar endpoint")
	w.Write([]byte("OK"))
}

func oneMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print("Executing middleware - 1")
		next.ServeHTTP(w, r)
		log.Print("Executing middleware - 2")
	})
}

func twoMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print("Executing middleware2 - 1")
		next.ServeHTTP(w, r)
		log.Print("Executing middleware2 - 2")
	})
}

func main() {
	mux := http.NewServeMux()

	// To apply the middleware across a particular path (/bar endpoint in our case)
	// mux.Handle("GET /foo", http.HandlerFunc(fooHandle))
	// mux.Handle("GET /bar", oneMiddleWare(http.HandlerFunc(barHandle)))

	// s := &http.Server{
	// 	Addr:    ":3000",
	// 	Handler: mux,
	// }

	// Using middleware on all routes
	mux.Handle("GET /foo", http.HandlerFunc(fooHandle))
	// Chaining different middlewares (one and two)
	mux.Handle("GET /bar", twoMiddleWare(http.HandlerFunc(barHandle)))

	s := http.Server{
		Addr:    ":3000",
		Handler: oneMiddleWare(mux),
	}

	err := s.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to listen at port 3000. Error: %v", err)
	}
}
