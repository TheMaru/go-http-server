package main

import (
	"log"
	"net/http"
)

func main() {
	const port = "8080"

	mux := http.NewServeMux()
	apiCfg := apiConfig{}

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	fs := http.FileServer(http.Dir("."))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fs)))
	mux.HandleFunc("/healthz", healthzHandler)
	mux.HandleFunc("/metrics", apiCfg.fileserverHitsHandler)
	mux.HandleFunc("/reset", apiCfg.resetHitsHandler)

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
