package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/TheMaru/go-http-server/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	const port = "8080"

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("db err: %v", err)
		log.Fatal("no connection to db")
	}

	mux := http.NewServeMux()
	apiCfg := apiConfig{
		dbQueries: database.New(db),
	}

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	fs := http.FileServer(http.Dir("."))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fs)))

	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("POST /api/chirps", apiCfg.createChirpHandler)
	mux.HandleFunc("POST /api/users", apiCfg.addUserHandler)

	mux.HandleFunc("POST /admin/reset", apiCfg.resetHitsHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
