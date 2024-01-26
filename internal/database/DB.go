package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/golang-jwt/jwt/v5"
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
	Email    string `json:"email"`
	Password []byte `json:"Password"`
}

type JWTUser struct {
	token string
	user  User
}

type DB struct {
	path string
	mux  *sync.RWMutex
}
type DBStructure struct {
	Chirps    map[int]Chirp           `json:"chirps"`
	Users     map[int]User            `json:"users"`
	Passwords map[string]UserPassword `json:"passwords"`
	JWTs      map[string]JWTUser      `json:"jwts"`
}

type loginResponse struct {
	id    int
	email string
	token string
}

func (db *DB) Login(password, email, jwtSecret string, expires_in_seconds int) (loginResponse, error) {
	db.ensureDB()
	data, err := db.loadDB()
	if err != nil {
		return loginResponse{}, err
	}

	ok := bcrypt.CompareHashAndPassword(data.Passwords[email].Password, []byte(password))

	if ok != nil {
		return loginResponse{}, ok
	}

	//Handle generating JWT

	// can't look at this
	for _, v := range data.Users {
		if v.Email == email {

			signedToken, err := generateSignedToken(expires_in_seconds, v.ID, jwtSecret)
			if err != nil {
				log.Fatal(err.Error())
			}
			data.JWTs[signedToken] = JWTUser{
				token: signedToken,
				user:  v,
			}
			response := loginResponse{
				v.ID,
				v.Email,
				signedToken,
			}
			return response, nil
		}
	}
	//this shouldn't happen at this point
	return loginResponse{}, errors.New("Failed to find user in db")
}

func generateSignedToken(expires_in_seconds, userID int, jwtSecret string) (string, error) {

	var expirationTime time.Time

	if expires_in_seconds > 0 && expires_in_seconds < 24 {
		expirationTime = time.Now().Add(time.Duration(expires_in_seconds))
	} else {
		expirationTime = time.Now().Add(time.Duration(time.Now().UTC().Day()))
	}
	newJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		Subject:   strconv.Itoa(userID),
	})
	log.Printf("jwtSecret: %s", jwtSecret)
	signedToken, err := newJWT.SignedString(jwtSecret)

	if err != nil {
		log.Print(err)
		return "", errors.New("Error generating signed token")
	}

	return signedToken, nil
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
		Chirps:    make(map[int]Chirp),
		Users:     make(map[int]User),
		Passwords: make(map[string]UserPassword),
		JWTs:      make(map[string]JWTUser),
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
	for _, v := range dbStructure.JWTs {
		resDBStructure.JWTs[v.token] = v
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
