package database

import (
	"errors"
	"strconv"
)

type Chirp struct {
	ID       int    `json:"id"`
	Body     string `json:"body"`
	AuthorID int    `json:"author_id"`
}

func (db *DB) CreateChirp(body string, ID string) (Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	if len(body) > 140 {
		return Chirp{}, errors.New("Chirp is too long")
	}

	authorID, err := strconv.Atoi(ID)
	if err != nil {
		panic(err)
	}

	chirp := Chirp{
		ID:       db.nextID,
		Body:     body,
		AuthorID: authorID,
	}

	db.chirps[chirp.ID] = chirp
	db.nextID++

	if err := db.writeDB(); err != nil {
		return Chirp{}, err
	}

	return chirp, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	chirps := make([]Chirp, 0, len(db.chirps))
	for _, chirp := range db.chirps {
		chirps = append(chirps, chirp)
	}

	return chirps, nil
}

func (db *DB) DeleteChirp(chirp Chirp) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	delete(db.chirps, chirp.ID)

	return nil
}
