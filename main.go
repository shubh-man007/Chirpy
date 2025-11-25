package main

import (
	"fmt"
	"time"
)

type Request struct {
	Path string
}

func handleRequest(req Request) {
	fmt.Printf("Processing %s\n", req.Path)
	time.Sleep(2 * time.Second)
	fmt.Printf("Processed %s\n", req.Path)
}

func handleRequests(reqs <-chan Request) {
	for req := range reqs {
		fmt.Printf(">> %s\n", req)
		go handleRequest(req)
	}
}

func main() {
	reqs := make(chan Request, 100)
	go handleRequests(reqs)
	for i := 0; i < 10; i++ {
		reqs <- Request{Path: fmt.Sprintf("/path/%d", i)}
		time.Sleep(500 * time.Millisecond)
	}

	time.Sleep(5 * time.Second)
	fmt.Println("5 seconds passed, killing server")
}
