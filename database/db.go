package database

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"sync"
)

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

type DB struct {
	path       string
	mux        *sync.RWMutex
	chirps     map[int]Chirp
	users      map[int]User
	nextID     int
	nextUserID int
	dbLoaded   bool
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

func NewDB(path string) (*DB, error) {
	db := &DB{
		path:       path,
		mux:        &sync.RWMutex{},
		chirps:     make(map[int]Chirp),
		users:      make(map[int]User),
		nextID:     1,
		nextUserID: 1,
		dbLoaded:   false,
	}

	if err := db.loadDB(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	if len(body) > 140 {
		return Chirp{}, errors.New("Chirp is too long")
	}

	chirp := Chirp{
		ID:   db.nextID,
		Body: body,
	}

	db.chirps[chirp.ID] = chirp
	db.nextID++

	if err := db.writeDB(); err != nil {
		return Chirp{}, err
	}

	return chirp, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	chirps := make([]Chirp, 0, len(db.chirps))
	for _, chirp := range db.chirps {
		chirps = append(chirps, chirp)
	}

	return chirps, nil
}

func (db *DB) CreateUser(email string) (User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	user := User{
		ID:    db.nextUserID,
		Email: email,
	}

	db.users[user.ID] = user
	db.nextUserID++

	if err := db.writeDB(); err != nil {
		return User{}, err
	}

	return user, nil
}

func (db *DB) writeDB() error {
	data, err := json.Marshal(map[string]interface{}{
		"chirps": db.chirps,
		"users":  db.users,
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

			body, ok := chirpMap["Body"].(string)
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

			user := User{
				ID:    id,
				Email: email,
			}
			db.users[id] = user
		}
		db.nextUserID = findMaxID(db.users) + 1
	}

	db.dbLoaded = true

	return nil
}

func (db *DB) DeleteDB(path string) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	db.chirps = make(map[int]Chirp)
	db.users = make(map[int]User)
	db.nextID = 1
	db.nextUserID = 1
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
	}

	return maxID
}

func (db *DB) Close() error {
	return db.writeDB()
}
