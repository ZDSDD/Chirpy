package database

import (
	"errors"
	"fmt"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

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

func (db *DB) CreateUser(email string, password string) (User, error) {
	db.ensureDB()
	data, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	if UserPass, keyAlreadyExists := data.Passwords[email]; keyAlreadyExists {
		//Return if the email is already in use.
		return User{}, errors.New(fmt.Sprintf("User [ %s ] alreadt exists.", UserPass.Email))
	}

	//If email is good, continue

	newUser := User{
		Email: email,
		ID:    len(data.Users) + 1,
	}

	encrypedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)

	newUserPasswd := UserPassword{
		Email:    email,
		Password: encrypedPassword,
	}

	// Update the in-memory DBStructure with the new chirp
	data.Users[newUser.ID] = newUser
	data.Passwords[newUser.Email] = newUserPasswd
	// Write the updated data back to the file
	err = db.writeDB(data)
	if err != nil {
		return User{}, err
	}

	return newUser, nil

}
