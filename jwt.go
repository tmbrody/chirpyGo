package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tmbrody/chirpyGo/database"
)

func (cfg *apiConfig) updateUserHandler(w http.ResponseWriter, r *http.Request) {
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

	userID, ok := token.Claims.(jwt.MapClaims)["sub"].(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Failed to extract user ID from JWT claims")
		return
	}

	updatedUser, err := db.UpdateUser(userID, params.Email, params.Password)
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

func extractJWTTokenFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}

func parseAndValidateJWTToken(cfg *apiConfig, tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		return []byte(cfg.jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}