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
	"fmt"
	"log"
	"net/http"
	"time"
)

const message = `
<html>
<head>
	<title>About</title>
</head>
<body>
	<p><stronig>Hi, <i>waddup</i> !</strong> This is a page about <b>SSE</b><p>
</body>
</html>
`

type ModelRes struct {
	output []string
}

func (tk *ModelRes) streamResHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(200)

	for _, val := range tk.output {
		content := fmt.Sprintf("data: %s\n\n", string(val))
		w.Write([]byte(content))
		w.(http.Flusher).Flush()

		time.Sleep(time.Millisecond * 500)
	}
}

func main() {
	mux := http.NewServeMux()

	tk := ModelRes{
		output: []string{"Am", "I", "a", "bloody", "AI", "generated", "response", "?"},
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "events.html")
	})

	mux.HandleFunc("/events", tk.streamResHandler)

	mux.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(200)

		_, err := w.Write([]byte(message))
		if err != nil {
			log.Print("Could not write to about")
		}
	})

	s := &http.Server{
		Addr:    ":5000",
		Handler: mux,
	}

	log.Print("Listening at port :5000")
	err := s.ListenAndServe()
	if err != nil {
		log.Printf("Could not listen to server at port: 5000. Err: %v", err)
	}
}
