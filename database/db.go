package database

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type DB struct {
	path     string
	mux      *sync.RWMutex
	chirps   map[int]Chirp
	nextID   int
	dbLoaded bool
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

func NewDB(path string) (*DB, error) {
	db := &DB{
		path:     path,
		mux:      &sync.RWMutex{},
		chirps:   make(map[int]Chirp),
		nextID:   1,
		dbLoaded: false,
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

	var dbStructure map[string]map[int]Chirp
	if err := json.Unmarshal(data, &dbStructure); err != nil && len(data) != 0 {
		return err
	}

	if dbStructure == nil || dbStructure["chirps"] == nil {
		dbStructure = map[string]map[int]Chirp{
			"chirps": make(map[int]Chirp),
		}
	}

	db.chirps = dbStructure["chirps"]
	db.nextID = findMaxID(db.chirps) + 1
	db.dbLoaded = true

	return nil
}

func (db *DB) writeDB() error {
	data, err := json.Marshal(map[string]interface{}{
		"chirps": db.chirps,
	})
	if err != nil {
		return err
	}

	return os.WriteFile(db.path, data, 0644)
}

func findMaxID(chirps map[int]Chirp) int {
	maxID := 0
	for id := range chirps {
		if id > maxID {
			maxID = id
		}
	}
	return maxID
}

func (db *DB) Close() error {
	return db.writeDB()
}
