package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type apiConfig struct {
	fileserverHits int
}

func main() {
	const filepathRoot = "."
	const port = "8080"
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

	r_admin.Get("/metrics", apiCfg.requestCounterHandler)

	r.Mount("/api", r_endpoints)
	r.Mount("/admin", r_admin)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
