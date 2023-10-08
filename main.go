package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tmbrody/chirpyGo/database"
)

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type apiConfig struct {
	fileserverHits int
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	db, err := database.NewDB("database.json")
	if err != nil {
		log.Fatalf("Error initializing the database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing the database: %v", err)
		}
	}()

	dbg := flag.Bool("debug", false, "Enable debug mode")

	flag.Parse()

	if *dbg {
		err := db.DeleteDB("database.json")
		if err != nil {
			log.Fatalf("Error deleting the database: %v", err)
		}
	}

	var apiCfg apiConfig

	r := chi.NewRouter()
	r_endpoints := chi.NewRouter()
	r_admin := chi.NewRouter()

	fileServer := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))

	r.Handle("/app/*", apiCfg.middlewareMetricsInc(fileServer))
	r.Handle("/app", apiCfg.middlewareMetricsInc(fileServer))

	corsMux := middlewareCors(r)

	r_endpoints.Get("/healthz", readinessHandler)
	r_endpoints.Get("/reset", apiCfg.resetCounterHandler)

	r_endpoints.Post("/chirps", withDB(createChirpHandler, db))
	r_endpoints.Get("/chirps", withDB(listChirpsHandler, db))
	r_endpoints.Get("/chirps/{chirpID}", withDB(GetChirpByID, db))

	r_endpoints.Post("/users", withDB(createUserHandler, db))

	r_admin.Get("/metrics", apiCfg.requestCounterHandler)

	r.Mount("/api", r_endpoints)
	r.Mount("/admin", r_admin)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)

	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func withDB(next http.HandlerFunc, db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "db", db)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
