package main

import (
	"encoding/json"
	"net/http"

	"github.com/tmbrody/chirpyGo/database"
)

func (cfg *apiConfig) polkaWebhookHandler(w http.ResponseWriter, r *http.Request) {
	tokenString := extractJWTTokenFromHeader(r)
	if tokenString == "" {
		respondWithError(w, http.StatusUnauthorized, "JWT token is missing or invalid")
		return
	}

	if tokenString != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "Invalid API key")
		return
	}

	ctx := r.Context()
	db, _ := ctx.Value(dbContextKey).(*database.DB)
	users, _ := db.GetUsers()

	var params struct {
		Event string `json:"event"`
		Data  struct {
			UserID int `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if params.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusOK, nil)
		return
	} else {
		if params.Data.UserID > len(users) {
			respondWithError(w, http.StatusBadRequest, "Invalid User ID")
			return
		}

		user := users[params.Data.UserID-1]

		_, err := db.UpdateUser(user.ID, user.Email, user.Password, true, true)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to update user data")
			return
		}

		respondWithJSON(w, http.StatusOK, nil)
		return
	}
}
