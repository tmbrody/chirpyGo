package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/tmbrody/chirpyGo/database"
)

var chirpsMutex sync.RWMutex

func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db, _ := ctx.Value(dbContextKey).(*database.DB)

	tokenString := extractJWTTokenFromHeader(r)
	if tokenString == "" {
		respondWithError(w, http.StatusUnauthorized, "JWT token is missing or invalid")
		return
	}

	token, err := parseAndValidateJWTToken(cfg, tokenString)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid JWT token")
		return
	}

	authorID, err := token.Claims.GetSubject()
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to find User ID")
	}

	var params struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	params.Body = remove_profanity(params.Body)

	chirp, err := db.CreateChirp(params.Body, authorID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create chirp")
		return
	}

	respondWithJSON(w, http.StatusCreated, chirp)
}

func listChirpsHandler(w http.ResponseWriter, r *http.Request) {
	chirpsMutex.Lock()
	defer chirpsMutex.Unlock()

	ctx := r.Context()
	db, _ := ctx.Value(dbContextKey).(*database.DB)

	chirps, err := db.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to find chirps")
		return
	}

	sort := r.URL.Query().Get("sort")
	id := r.URL.Query().Get("author_id")

	var response []database.Chirp

	if sort != "desc" {
		sort = "asc"
	}

	if id == "" {
		if sort == "asc" {
			respondWithJSON(w, http.StatusOK, chirps)
			return
		} else if sort == "desc" {
			for i := len(chirps) - 1; i >= 0; i-- {
				response = append(response, chirps[i])
			}

			respondWithJSON(w, http.StatusOK, response)
			return
		}
	}

	authorID, err := strconv.Atoi(id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to convert author ID into int")
		return
	}

	if sort == "asc" {
		for _, chirp := range chirps {
			if chirp.AuthorID == authorID {
				response = append(response, chirp)
			}
		}
	} else if sort == "desc" {
		for i := len(chirps) - 1; i >= 0; i-- {
			if chirps[i].AuthorID == authorID {
				response = append(response, chirps[i])
			}
		}
	}

	respondWithJSON(w, http.StatusOK, response)
}

func getChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpString := chi.URLParam(r, "chirpID")

	chirpID, err := strconv.Atoi(chirpString)
	if err != nil {
		http.Error(w, "Invalid chirp ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	db, _ := ctx.Value(dbContextKey).(*database.DB)
	chirps, err := db.GetChirps()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch chirps")
		return
	}

	if chirpID <= 0 || chirpID > len(chirps) {
		respondWithError(w, http.StatusNotFound, "Chirp not found")
	} else {
		respondWithJSON(w, http.StatusOK, chirps[chirpID-1])
	}
}

func (cfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db, _ := ctx.Value(dbContextKey).(*database.DB)

	chirpString := chi.URLParam(r, "chirpID")

	chirpID, err := strconv.Atoi(chirpString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}

	chirps, err := db.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to find chirps")
		return
	}

	tokenString := extractJWTTokenFromHeader(r)
	if tokenString == "" {
		respondWithError(w, http.StatusUnauthorized, "JWT token is missing or invalid")
		return
	}

	token, err := parseAndValidateJWTToken(cfg, tokenString)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid JWT token")
		return
	}

	authorIDString, err := token.Claims.GetSubject()
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to find Author ID")
		return
	}

	authorID, err := strconv.Atoi(authorIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid author ID")
		return
	}

	if chirpID != authorID {
		respondWithError(w, http.StatusForbidden, "Can't delete chirp from different account")
		return
	}

	err = db.DeleteChirp(chirps[chirpID-1])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to delete chirp")
		return
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
