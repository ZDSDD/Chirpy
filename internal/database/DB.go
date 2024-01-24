package database

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"golang.org/x/crypto/bcrypt"
	"errors"
)

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

type UserPassword struct {
	Email string `json:"email"`
	Password []byte `json:"Password"`
}

type DB struct {
	path string
	mux  *sync.RWMutex
}
type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
	Passwords map[string]UserPassword `json:"passwords"`
}


func (db *DB) Login(password, email string) (User, error){
	db.ensureDB()
	data, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	ok := bcrypt.CompareHashAndPassword(data.Passwords[email].Password,[]byte(password))

	if ok != nil{
		return User{},ok
	}

	for _, v := range data.Users {
		if v.Email == email{
			return v,nil
		}
	}
	//this shouldn't happen at this point
	return User{}, errors.New("Failed to find user in db")
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
		Passwords: make(map[string]UserPassword),
	}

	for _, v := range dbStructure.Chirps {
		resDBStructure.Chirps[v.ID] = v
	}
	for _, v := range dbStructure.Users {
		resDBStructure.Users[v.ID] = v
	}
	for _, v := range dbStructure.Passwords {
		resDBStructure.Passwords[v.Email] = v
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
