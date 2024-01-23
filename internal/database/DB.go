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

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

type DB struct {
	path string
	mux  *sync.RWMutex
}
type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
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
	// Write the updated data back to the file
	err = db.writeDB(data)
	if err != nil {
		return Chirp{}, err
	}

	return newChirp, nil
}

func (db *DB) CreateUser(email string) (User, error) {
	db.ensureDB()
	data, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	newUser := User{
		Email: email,
		ID:    len(data.Users) + 1,
	}

	// Update the in-memory DBStructure with the new chirp
	data.Users[newUser.ID] = newUser
	// Write the updated data back to the file
	err = db.writeDB(data)
	if err != nil {
		return User{}, err
	}

	return newUser, nil

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
	allDataRaw, err := os.ReadFile(db.path)
	if err != nil {
		return DBStructure{}, err
	}

	var dbStructure DBStructure
	err = json.Unmarshal(allDataRaw, &dbStructure)
	if err != nil {
		log.Printf("error decoding allChirpsRaw: %v", err)
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}
		log.Printf("sakura response: %q", allDataRaw)
		return DBStructure{}, err
	}

	resDBStructure := DBStructure{
		Chirps: make(map[int]Chirp),
		Users:  make(map[int]User),
	}

	for _, v := range dbStructure.Chirps {
		resDBStructure.Chirps[v.ID] = v
	}
	for _, v := range dbStructure.Users {
		resDBStructure.Users[v.ID] = v
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
