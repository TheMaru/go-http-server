package main

import (
	"log"
	"net/http"
)

func main() {
	const port = "8080"

	mux := http.NewServeMux()

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	fs := http.FileServer(http.Dir("."))
	mux.Handle("/app/", http.StripPrefix("/app", fs))
	mux.HandleFunc("/healthz", healthzHandler)

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
