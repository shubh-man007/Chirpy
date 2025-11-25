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
// 		fmt.Printf(">> %s\n", req)
// 		go handleRequest(req)
// 	}
// }

// func main() {
// 	reqs := make(chan Request, 100)
// 	go handleRequests(reqs)
// 	for i := 0; i < 10; i++ {
// 		reqs <- Request{Path: fmt.Sprintf("/path/%d", i)}
// 		time.Sleep(500 * time.Millisecond)
// 	}

// 	time.Sleep(5 * time.Second)
// 	fmt.Println("5 seconds passed, killing server")
// }

package main

import (
	"fmt"
	"time"
)

func handleRequests(reqs <-chan request) {
	for req := range reqs {
		go handleRequest(req)
	}
}

// don't touch below this line

type request struct {
	path string
}

func main() {
	reqs := make(chan request, 100)
	go handleRequests(reqs)
	for i := 0; i < 10; i++ {
		reqs <- request{path: fmt.Sprintf("/path/%d", i)}
		time.Sleep(500 * time.Millisecond)
	}

	time.Sleep(5 * time.Second)
	fmt.Println("5 seconds passed, killing server")
}

func handleRequest(req request) {
	fmt.Println("Handling request for", req.path)
	time.Sleep(2 * time.Second)
	fmt.Println("Done with request for", req.path)
}
