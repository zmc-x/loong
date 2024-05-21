package main

import (
    "fmt"
    "net/http"
    "github.com/gorilla/mux"
)

func pingHandler(port string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "pong from port %s", port)
    }
}

func searchHandler(port string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "search endpoint from port %s", port)
    }
}

func startServer(addr string, handler http.Handler) {
    fmt.Printf("Starting server on %s\n", addr)
    if err := http.ListenAndServe(addr, handler); err != nil {
        fmt.Printf("Error starting server on %s: %v\n", addr, err)
    }
}

func main() {
    // Create routers with handlers that include port information
    pingRouter1 := mux.NewRouter()
    pingRouter1.HandleFunc("/ping", pingHandler("9095")).Methods("GET")

    pingRouter2 := mux.NewRouter()
    pingRouter2.HandleFunc("/ping", pingHandler("9096")).Methods("GET")

    searchRouter1 := mux.NewRouter()
    searchRouter1.HandleFunc("/search", searchHandler("9097")).Methods("GET")

    searchRouter2 := mux.NewRouter()
    searchRouter2.HandleFunc("/search", searchHandler("9098")).Methods("GET")

    // Start servers
    go startServer(":9095", pingRouter1)
    go startServer(":9096", pingRouter2)
    go startServer(":9097", searchRouter1)
    go startServer(":9098", searchRouter2)

    // Prevent main from exiting
    select {}
}
