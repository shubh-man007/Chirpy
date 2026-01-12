// // package main

// // import (
// // 	"fmt"
// // 	"time"
// // )

// // type Request struct {
// // 	Path string
// // }

// // func handleRequest(req Request) {
// // 	fmt.Printf("Processing %s\n", req.Path)
// // 	time.Sleep(2 * time.Second)
// // 	fmt.Printf("Processed %s\n", req.Path)
// // }

// // func handleRequests(reqs <-chan Request) {
// // 	for req := range reqs {
// // 		go handleRequest(req)
// // 	}
// // }

// // func main() {
// // 	reqs := make(chan Request, 100)
// // 	go handleRequests(reqs)
// // 	for i := 0; i < 4; i++ {
// // 		reqs <- Request{Path: fmt.Sprintf("/path/%d", i)}
// // 		time.Sleep(500 * time.Millisecond)
// // 	}

// // 	time.Sleep(5 * time.Second)
// // 	fmt.Println("5 seconds passed, killing server")
// // }

// package main

// import (
// 	"log"
// 	"net/http"
// )

// func fooHandle(w http.ResponseWriter, r *http.Request) {
// 	log.Print("Executing foo endpoint")
// 	w.Write([]byte("OK"))
// }

// func barHandle(w http.ResponseWriter, r *http.Request) {
// 	log.Print("Executing bar endpoint")
// 	w.Write([]byte("OK"))
// }

// func oneMiddleWare(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		log.Print("Executing middleware - 1")
// 		next.ServeHTTP(w, r)
// 		log.Print("Executing middleware - 2")
// 	})
// }

// func twoMiddleWare(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		log.Print("Executing middleware2 - 1")
// 		next.ServeHTTP(w, r)
// 		log.Print("Executing middleware2 - 2")
// 	})
// }

// func main() {
// 	mux := http.NewServeMux()

// 	// To apply the middleware across a particular path (/bar endpoint in our case)
// 	// mux.Handle("GET /foo", http.HandlerFunc(fooHandle))
// 	// mux.Handle("GET /bar", oneMiddleWare(http.HandlerFunc(barHandle)))

// 	// s := &http.Server{
// 	// 	Addr:    ":3000",
// 	// 	Handler: mux,
// 	// }

// 	// Using middleware on all routes
// 	mux.Handle("GET /foo", http.HandlerFunc(fooHandle))
// 	// Chaining different middlewares (one and two)
// 	mux.Handle("GET /bar", twoMiddleWare(http.HandlerFunc(barHandle)))

// 	s := http.Server{
// 		Addr:    ":3000",
// 		Handler: oneMiddleWare(mux),
// 	}

// 	err := s.ListenAndServe()
// 	if err != nil {
// 		log.Fatalf("Failed to listen at port 3000. Error: %v", err)
// 	}
// }

package main

import (
	"context"
	"fmt"
	"time"
)

func watch(ctx context.Context, name string) {
	for {
		select {
		case <-ctx.Done():
			// After the main goroutine calls cancel, a signal will be sent to the ctx.Done() channel, and this part will receive the message
			fmt.Printf("%s exit!\n", name)
			return
		default:
			fmt.Printf("%s working...\n", name)
			time.Sleep(2 * time.Second)
		}
	}
}

func valFunc(ctx context.Context) {
	fmt.Printf("Name: %s\n", ctx.Value("name").(string))
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	ctx1 := context.WithValue(context.Background(), "name", "Shubh")

	go watch(ctx, "Go1")
	go watch(ctx, "Go2")
	go valFunc(ctx1)

	time.Sleep(6 * time.Second)
	fmt.Println("[CANCEL]")
	cancel() // Notify goroutine1 and goroutine2 to close
	time.Sleep(1 * time.Second)
}
