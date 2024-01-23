package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
)

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type DB struct {
	path string
	mux  *sync.RWMutex
}
type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	newDB := DB{
		path: path,
		mux:  &sync.RWMutex{},
	}
	err := newDB.ensureDB()
	if err != nil {
		return nil, err
	}
	return &newDB, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (Chirp, error) {
	db.ensureDB()
	data, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	newChirp := Chirp{
		Body: body,
		ID:   len(data.Chirps) + 1,
	}

	// Update the in-memory DBStructure with the new chirp
	data.Chirps[newChirp.ID] = newChirp
	log.Printf("%v", data.Chirps[newChirp.ID])
	// Write the updated data back to the file
	err = db.writeDB(data)
	if err != nil {
		return Chirp{}, err
	}

	return newChirp, nil
}

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

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	if !checkFileExists(db.path) {
		_, err := os.Create(db.path)
		if err != nil {
			log.Fatal(err)
			return err
		}
		db.writeDB(DBStructure{})
	}
	return nil
}

func checkFileExists(filePath string) bool {
	_, error := os.Stat(filePath)
	//return !os.IsNotExist(err)
	return !errors.Is(error, os.ErrNotExist)
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	allChirpsRaw, err := os.ReadFile(db.path)
	if err != nil {
		return DBStructure{}, err
	}

	var allChirps DBStructure
	err = json.Unmarshal(allChirpsRaw, &allChirps)
	if err != nil {
		log.Printf("error decoding sakura response: %v", err)
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}
		log.Printf("sakura response: %q", allChirpsRaw)
		return DBStructure{}, err
	}

	resDBStructure := DBStructure{
		Chirps: make(map[int]Chirp),
	}

	for _, v := range allChirps.Chirps {
		resDBStructure.Chirps[v.ID] = v
	}

	return resDBStructure, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	// Open the file in write mode, truncating it if it exists, create it if it doesn't exist
	file, err := os.OpenFile(db.path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Data to write
	data, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}

	// Write the data to the file
	n, err := file.Write(data)
	if err != nil {
		return err
	}

	fmt.Printf("Written: %d bytes", n)

	return nil
}
