package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) resetCounterHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "Hits have been reset to 0.")
}
