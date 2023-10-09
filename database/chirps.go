package database

import (
	"errors"
)

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
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
