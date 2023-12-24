package main

import (
    "fmt"
    "io"
    "net/http"
    "os"
    "github.com/gorilla/mux"
    "log"
)

var (
    ntfy_url = os.Getenv("NTFY_URL")
)

func getHealthCheck(w http.ResponseWriter, r *http.Request) {
    io.WriteString(w, "OK!")
}

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/health", getHealthCheck)

    port := os.Getenv("NTFY_ATOM_PORT")
    if (port == "") {
        port = "8080"
    }

    if (ntfy_url == "") {
        log.Fatalf("NTFY_URL not set, cannot start")
    }

    log.Println("listen on", port)
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
}
