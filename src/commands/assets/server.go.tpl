// main.go
package main

import (
	"fmt"
	"net/http"
	"{{ .Module }}/internal/greeting"
)

var (
	name        string
	description string
	version     string
	buildTime   string
)

func main() {
	// Define a handler function to handle incoming HTTP requests
	handler := func(w http.ResponseWriter, req *http.Request) {
		// Use the greeting package to format the response
		greeting.Format(w, name, version, description, buildTime)
	}

	// Register the handler function to handle all requests to the "/" path
	http.HandleFunc("/", handler)

	// Start the HTTP server on port 8080
	fmt.Println("Server is listening on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
