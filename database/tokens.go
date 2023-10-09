package database

type RevokedToken struct {
	ID             int    `json:"id"`
	RevokedTokenID string `json:"revoked_token_id"`
}

func (db *DB) StoreRevokedToken(tokenID string) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	revokedToken := RevokedToken{
		ID:             db.nextRevokedTokenID,
		RevokedTokenID: tokenID,
	}

	db.revokedTokens[revokedToken.ID] = revokedToken
	db.nextRevokedTokenID++

	if err := db.writeDB(); err != nil {
		return err
	}

	return nil
}

func (db *DB) GetRevokedTokens() ([]RevokedToken, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	revokedTokens := make([]RevokedToken, 0, len(db.revokedTokens))
	for _, revokedToken := range db.revokedTokens {
		revokedTokens = append(revokedTokens, revokedToken)
	}

	return revokedTokens, nil
}
