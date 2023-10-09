package database

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"sync"
)

type DB struct {
	path               string
	mux                *sync.RWMutex
	chirps             map[int]Chirp
	users              map[int]User
	revokedTokens      map[int]RevokedToken
	nextID             int
	nextUserID         int
	nextRevokedTokenID int
	dbLoaded           bool
}

type DBStructure struct {
	Chirps        map[int]Chirp        `json:"chirps"`
	Users         map[int]User         `json:"users"`
	RevokedTokens map[int]RevokedToken `json:"revoked_tokens"`
}

func NewDB(path string) (*DB, error) {
	db := &DB{
		path:               path,
		mux:                &sync.RWMutex{},
		chirps:             make(map[int]Chirp),
		users:              make(map[int]User),
		revokedTokens:      make(map[int]RevokedToken),
		nextID:             1,
		nextUserID:         1,
		nextRevokedTokenID: 1,
		dbLoaded:           false,
	}

	if err := db.loadDB(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) writeDB() error {
	data, err := json.Marshal(map[string]interface{}{
		"chirps":         db.chirps,
		"users":          db.users,
		"revoked_tokens": db.revokedTokens,
	})
	if err != nil {
		return err
	}

	return os.WriteFile(db.path, data, 0644)
}

func (db *DB) loadDB() error {
	db.mux.Lock()
	defer db.mux.Unlock()

	if db.dbLoaded {
		return nil
	}

	data, err := os.ReadFile(db.path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var dbStructure map[string]interface{}
	if err := json.Unmarshal(data, &dbStructure); err != nil && len(data) != 0 {
		return err
	}

	if dbStructure == nil {
		dbStructure = make(map[string]interface{})
	}

	if chirpsData, ok := dbStructure["chirps"].(map[string]interface{}); ok {
		db.chirps = make(map[int]Chirp)
		for key, val := range chirpsData {
			id, _ := strconv.Atoi(key)

			chirpMap, ok := val.(map[string]interface{})
			if !ok {
				return errors.New("chirp data is invalid")
			}

			body, ok := chirpMap["body"].(string)
			if !ok {
				return errors.New("chirp body is missing or not a string")
			}

			chirp := Chirp{
				ID:   id,
				Body: body,
			}
			db.chirps[id] = chirp
		}
		db.nextID = findMaxID(db.chirps) + 1
	}

	if usersData, ok := dbStructure["users"].(map[string]interface{}); ok {
		db.users = make(map[int]User)
		for key, val := range usersData {
			id, _ := strconv.Atoi(key)

			userMap, ok := val.(map[string]interface{})
			if !ok {
				return errors.New("user data is invalid")
			}

			email, ok := userMap["email"].(string)
			if !ok {
				return errors.New("user email is missing or not a string")
			}

			password, ok := userMap["password"].(string)
			if !ok {
				return errors.New("user password is missing or not a string")
			}

			user := User{
				ID:       id,
				Email:    email,
				Password: password,
			}
			db.users[id] = user
		}
		db.nextUserID = findMaxID(db.users) + 1
	}

	if revokedTokensData, ok := dbStructure["revoked_tokens"].(map[string]interface{}); ok {
		db.revokedTokens = make(map[int]RevokedToken)
		for key, val := range revokedTokensData {
			id, _ := strconv.Atoi(key)

			revokedTokenMap, ok := val.(map[string]interface{})
			if !ok {
				return errors.New("user data is invalid")
			}

			revokedTokenID, ok := revokedTokenMap["revoked_token_id"].(string)
			if !ok {
				return errors.New("revoked token id is missing")
			}

			revokedToken := RevokedToken{
				ID:             id,
				RevokedTokenID: revokedTokenID,
			}
			db.revokedTokens[id] = revokedToken
		}
		db.nextRevokedTokenID = findMaxID(db.revokedTokens) + 1
	}

	db.dbLoaded = true

	return nil
}

func (db *DB) DeleteDB(path string) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	db.chirps = make(map[int]Chirp)
	db.users = make(map[int]User)
	db.revokedTokens = make(map[int]RevokedToken)
	db.nextID = 1
	db.nextUserID = 1
	db.nextRevokedTokenID = 1
	db.dbLoaded = false

	err := os.Remove(path)

	if err != nil {
		return err
	}

	return nil
}

func findMaxID(jsonData interface{}) int {
	var maxID int

	switch v := jsonData.(type) {
	case map[int]Chirp:
		for id := range v {
			if id > maxID {
				maxID = id
			}
		}
	case map[int]User:
		for id := range v {
			if id > maxID {
				maxID = id
			}
		}
	case map[int]RevokedToken:
		for id := range v {
			if id > maxID {
				maxID = id
			}
		}
	}

	return maxID
}

func (db *DB) Close() error {
	return db.writeDB()
}
