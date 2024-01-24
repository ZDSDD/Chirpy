package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ZDSDD/Chirpy/internal/database"
	"github.com/go-chi/chi/v5"
)

type apiConfig struct {
	fileserverHits int
	localDB        *database.DB
	jwtSecret      string
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(fmt.Sprint("<h1>Welcome, Chirpy Admin</h1>")))
	w.Write([]byte(fmt.Sprintf("<p>Chirpy has been visited %d times!</p>", cfg.fileserverHits)))
	w.WriteHeader(http.StatusOK)
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
	w.WriteHeader(http.StatusOK)
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits += 1
		//w.Header().Set("Cache-Control", "no-cache")
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) postChirpHandler(w http.ResponseWriter, r *http.Request) {
	validatedChirp := database.Chirp{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&validatedChirp); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	err := validateChirpBody(validatedChirp.Body, w)

	if err != nil {
		log.Printf("error validating a chirp: %s", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	newChirp, err := cfg.localDB.CreateChirp(validatedChirp.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	err = respondWithJSON(w, 201, newChirp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
}

func (cfg *apiConfig) getChirpHandler(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.localDB.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sort.Slice(chirps, func(i, j int) bool { return chirps[i].ID < chirps[j].ID })
	respondWithJSON(w, 200, chirps)
}

func (cfg *apiConfig) getChirpByIDHandler(w http.ResponseWriter, r *http.Request) {
	chirpIDstring := chi.URLParam(r, "chirpID")
	chirpID, err := strconv.Atoi(chirpIDstring)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	chirp, err := cfg.localDB.GetChirp(chirpID)
	if err != nil {
		respondWithError(w, 404, err.Error())
		return
	}
	respondWithJSON(w, 200, chirp)
}

func (cfg *apiConfig) postUserHandler(w http.ResponseWriter, r *http.Request) {

	type params struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	requestBody := params{}
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&requestBody); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	createdUser, err := cfg.localDB.CreateUser(requestBody.Email, requestBody.Password)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = respondWithJSON(w, 201, createdUser)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

}

func (cfg *apiConfig) postLoginHandler(w http.ResponseWriter, r *http.Request) {

	type params struct {
		Password           string `json:"password"`
		Email              string `json:"email"`
		Expires_in_seconds int    `json:"expires_in_seconds"`
	}
	requestBody := params{}
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&requestBody); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	user, err := cfg.localDB.Login(requestBody.Password, requestBody.Email)

	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	var expirationTime time.Time

	if requestBody.Expires_in_seconds > 0 && requestBody.Expires_in_seconds < 24 {
		expirationTime = time.Now().Add(time.Duration(requestBody.Expires_in_seconds))
	} else {
		expirationTime = time.Now().Add(time.Duration(time.Now().UTC().Day()))
	}
	newJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		Subject:   strconv.Itoa(user.ID),
	})

	signedToken, err := newJWT.SignedString(cfg.jwtSecret)

	if err != nil {

	}

	response := struct {
		id    int
		email string
		token string
	}{
		user.ID,
		user.Email,
		signedToken,
	}

	err = respondWithJSON(w, 200, response)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
}
