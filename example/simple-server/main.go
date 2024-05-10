package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Create a new mux router
	r := mux.NewRouter()

	// Define routes
	r.HandleFunc("/route1", routeHandler("Route 1")).Methods("GET")
	r.HandleFunc("/route2", routeHandler("Route 2")).Methods("GET")
	r.HandleFunc("/route3", routeHandler("Route 3")).Methods("GET")

	// Run multiple servers concurrently on different ports
	go func() {
		fmt.Println("Server running on port 9095...")
		http.ListenAndServe(":9095", r)
	}()
	go func() {
		fmt.Println("Server running on port 9096...")
		http.ListenAndServe(":9096", r)
	}()
	go func() {
		fmt.Println("Server running on port 9097...")
		http.ListenAndServe(":9097", r)
	}()

	// Keep the main goroutine running
	select {}
}

// Handler function for all routes
func routeHandler(message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Message from %s", message)
	}
}
