package main

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, r *http.Request) {
	chirpIDString := chi.URLParam(r, "chirpID")
	chirpID, err := strconv.Atoi(chirpIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}

	dbChirp, err := cfg.DB.GetChirp(chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get chirp")
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		AuthorID: dbChirp.AuthorID,
		ID:       dbChirp.ID,
		Body:     dbChirp.Body,
	})
}

func (cfg *apiConfig) handlerChirpsRetrieve(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.DB.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
		return
	}

	authorID := -1

	s := r.URL.Query().Get("author_id")
	sortOrder := r.URL.Query().Get("sort")
	if sortOrder == "" {
		sortOrder = "asc"
	}

	if s != "" {
		authorID, err = strconv.Atoi(s)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
	}

	chirps := []Chirp{}

	// query param was empty
	if authorID == -1 {
		for _, dbChirp := range dbChirps {
			chirps = append(chirps, Chirp{
				AuthorID: dbChirp.AuthorID,
				ID:       dbChirp.ID,
				Body:     dbChirp.Body,
			})
		}
	} else { //query param wasn't empty
		for _, dbChirp := range dbChirps {
			if dbChirp.AuthorID != authorID {
				continue
			}
			chirps = append(chirps, Chirp{
				AuthorID: dbChirp.AuthorID,
				ID:       dbChirp.ID,
				Body:     dbChirp.Body,
			})
		}
	}
	if sortOrder == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].ID > chirps[j].ID
		})
	} else {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].ID < chirps[j].ID
		})
	}

	respondWithJSON(w, http.StatusOK, chirps)
}
