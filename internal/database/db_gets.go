package database

import (
	"errors"
	"fmt"
)


// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	var result []Chirp
	db.ensureDB()
	data, err := db.loadDB()
	if err != nil {
		return []Chirp{}, err
	}
	for _, v := range data.Chirps {
		result = append(result, v)
	}
	return result, nil
}

func (db *DB) GetChirp(id int) (Chirp, error) {
	db.ensureDB()
	data, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	result, ok := data.Chirps[id]

	if !ok {
		return Chirp{}, errors.New(fmt.Sprintf("No such key [%d]", id))
	}
	return result, nil
}

