package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/tmbrody/chirpyGo/database"
)

func createChirpHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db, _ := ctx.Value("db").(*database.DB)

	var params struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	params.Body = remove_profanity(params.Body)

	chirp, err := db.CreateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create chirp")
		return
	}

	respondWithJSON(w, http.StatusCreated, chirp)
}

func listChirpsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db, _ := ctx.Value("db").(*database.DB)
	chirps, err := db.GetChirps()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch chirps")
		return
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	response := fmt.Sprintf(`{"error": "%s"}`, msg)
	w.Write([]byte(response))
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Error encoding JSON: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
	}
}

func remove_profanity(original_body string) string {
	profane_words := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}

	words := strings.Split(original_body, " ")
	cleanWords := make([]string, len(words))

	for idx, word := range words {
		lc_word := strings.ToLower(word)

		if profane_words[lc_word] {
			cleanWords[idx] = "****"
		} else {
			cleanWords[idx] = word
		}
	}

	return strings.Join(cleanWords, " ")
}
