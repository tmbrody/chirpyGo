package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tmbrody/chirpyGo/database"
	"golang.org/x/crypto/bcrypt"
)

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db, _ := ctx.Value(dbContextKey).(*database.DB)

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

func (cfg *apiConfig) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db, _ := ctx.Value(dbContextKey).(*database.DB)
	users, _ := db.GetUsers()

	var params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON")
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

	issuer, _ := token.Claims.GetIssuer()
	if issuer == "chirpy-refresh" {
		respondWithError(w, http.StatusUnauthorized, "Using JWT refresh token")
		return
	}

	userID, ok := token.Claims.(jwt.MapClaims)["sub"].(int)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Failed to extract user ID from JWT claims")
		return
	}

	updatedUser, err := db.UpdateUser(userID, params.Email, params.Password, users[userID].IsChirpyRed, false)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update user data")
		return
	}

	response := map[string]interface{}{
		"id":    updatedUser.ID,
		"email": updatedUser.Email,
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (cfg *apiConfig) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db, _ := ctx.Value(dbContextKey).(*database.DB)
	users, err := db.GetUsers()

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
		if params.Email == user.Email {
			if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(params.Password)); err != nil {
				respondWithError(w, http.StatusUnauthorized, "Wrong password")
				return
			}

			accessClaims := jwt.RegisteredClaims{
				Issuer:   "chirpy-access",
				Subject:  strconv.Itoa(user.ID),
				IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
			}

			refreshClaims := jwt.RegisteredClaims{
				Issuer:   "chirpy-refresh",
				Subject:  strconv.Itoa(user.ID),
				IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
			}

			accessExpiration := time.Hour
			accessClaims.ExpiresAt = jwt.NewNumericDate(time.Now().UTC().Add(accessExpiration))

			refreshExpiration := 60 * (24 * time.Hour)
			refreshClaims.ExpiresAt = jwt.NewNumericDate(time.Now().UTC().Add(refreshExpiration))

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
			refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

			signedToken, err := token.SignedString([]byte(cfg.jwtSecret))
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Failed to generate token")
				return
			}

			signedRefreshToken, err := refreshToken.SignedString([]byte(cfg.jwtSecret))
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Failed to generate token")
				return
			}

			userID, _ := strconv.Atoi(accessClaims.Subject)

			response := map[string]interface{}{
				"id":            userID,
				"email":         user.Email,
				"token":         signedToken,
				"refresh_token": signedRefreshToken,
				"is_chirpy_red": user.IsChirpyRed,
			}

			respondWithJSON(w, http.StatusOK, response)
			return
		}
	}

	respondWithError(w, http.StatusBadRequest, "User not found")
}
