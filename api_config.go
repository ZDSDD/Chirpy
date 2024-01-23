package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"

	"github.com/ZDSDD/Chirpy/internal/database"
	"github.com/go-chi/chi/v5"
)

type apiConfig struct {
	fileserverHits int
	localDB        *database.DB
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
	user := database.User{}
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&user); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	createdUser, err := cfg.localDB.CreateUser(user.Email)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	err = respondWithJSON(w, 201, createdUser)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

}
