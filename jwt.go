package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tmbrody/chirpyGo/database"
)

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

func (cfg *apiConfig) refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
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

	issuer, _ := token.Claims.GetIssuer()
	if issuer != "chirpy-refresh" {
		respondWithError(w, http.StatusUnauthorized, "Not using JWT refresh token")
		return
	}

	revokedTokens, err := db.GetRevokedTokens()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch revoked tokens")
		return
	}

	for _, revokedToken := range revokedTokens {
		if tokenString == revokedToken.RevokedTokenID {
			respondWithError(w, http.StatusUnauthorized, "Using revoked JWT refresh token")
			return
		}
	}

	userID, err := token.Claims.GetSubject()
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to find User ID")
	}

	accessClaims := jwt.RegisteredClaims{
		Issuer:   "chirpy-access",
		Subject:  userID,
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
	}

	accessExpiration := time.Hour
	accessClaims.ExpiresAt = jwt.NewNumericDate(time.Now().UTC().Add(accessExpiration))

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)

	signedNewToken, err := newToken.SignedString([]byte(cfg.jwtSecret))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	response := map[string]interface{}{
		"token": signedNewToken,
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (cfg *apiConfig) revokeTokenHandler(w http.ResponseWriter, r *http.Request) {
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

	issuer, _ := token.Claims.GetIssuer()
	if issuer != "chirpy-refresh" {
		respondWithError(w, http.StatusUnauthorized, "Not using JWT refresh token")
		return
	}

	err = db.StoreRevokedToken(tokenString)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, "JWT refresh token successfully revoked")
}
