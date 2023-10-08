package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/tmbrody/chirpyGo/database"
	"golang.org/x/crypto/bcrypt"
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

func GetChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpString := chi.URLParam(r, "chirpID")

	chirpID, err := strconv.Atoi(chirpString)
	if err != nil {
		http.Error(w, "Invalid chirp ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	db, _ := ctx.Value("db").(*database.DB)
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

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db, _ := ctx.Value("db").(*database.DB)

	var params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	user, err := db.CreateUser(params.Email, params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, user)
}

func loginUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db, _ := ctx.Value("db").(*database.DB)
	users, usersResponse, err := db.GetUsers()

	var params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch users")
		return
	}

	for _, user := range users {
		for _, userResponse := range usersResponse {
			if params.Email == userResponse.Email {
				if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(params.Password)); err != nil {
					respondWithError(w, http.StatusUnauthorized, "Wrong password")
					return
				}

				respondWithJSON(w, http.StatusOK, userResponse)
				return
			}
		}
	}

	respondWithError(w, http.StatusBadRequest, "User not found")
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
