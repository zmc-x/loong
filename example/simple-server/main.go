package main

import (
    "fmt"
    "net/http"
    "github.com/gorilla/mux"
)

func Handler(path, port string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "request %s from port %s, methods is %s", path, port, r.Method)
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
    pingRouter1.HandleFunc("/ping", Handler("/ping", "9095"))
    pingRouter2 := mux.NewRouter()
    pingRouter2.HandleFunc("/ping", Handler("/ping", "9098"))

    pongRouter2 := mux.NewRouter()
    pongRouter2.HandleFunc("/pong", Handler("/pong", "9096"))

    searchRouter1 := mux.NewRouter()
    searchRouter1.HandleFunc("/search", Handler("/search", "9097"))

    // Start servers
    go startServer(":9095", pingRouter1)
    go startServer(":9096", pongRouter2)
    go startServer(":9097", searchRouter1)
    go startServer(":9098", pingRouter2)
    // Prevent main from exiting
    select {}
}
