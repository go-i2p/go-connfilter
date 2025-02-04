package main

import (
	"log"
	"net"
	"net/http"

	httpinspector "github.com/go-i2p/go-connfilter/http"
)

func main() {
	// Create a regular TCP listener
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	// Create inspector with custom configuration
	config := httpinspector.Config{
		OnRequest: func(req *http.Request) error {
			// Add custom header to all requests
			req.Header.Set("X-Inspected", "true")
			return nil
		},
		OnResponse: func(resp *http.Response) error {
			// Log all response status codes
			log.Printf("Response status: %s", resp.Status)
			return nil
		},
		LoggingEnabled: true,
	}

	inspector := httpinspector.New(listener, config)
	defer inspector.Close()

	// Use the inspector with http.Server
	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello, World!"))
		}),
	}

	log.Fatal(server.Serve(inspector))
}
