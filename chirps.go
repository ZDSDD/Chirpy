package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/ZDSDD/Chirpy/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	Body string `json:"body"`
}

func cleanProfaneWords(s string, profaneWords []string) string {
	words := strings.Fields(s)
	var sb strings.Builder
	for i, word := range words {
		if slices.Contains(profaneWords, strings.ToLower(word)) {
			sb.WriteString("****")
		} else {
			sb.WriteString(word)
		}
		if i < len(words)-1 {
			sb.WriteRune(' ')
		}
	}
	return sb.String()
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	var chirp = &Chirp{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(chirp)
	if err != nil {
		// an error will be thrown if the JSON is invalid or has the wrong types
		// any missing fields will simply have their values in the struct set to their zero value
		errorMsg := fmt.Sprintf("Error decoding parameters: %s", err)
		log.Printf(errorMsg)
		responseWithJsonError(w, errorMsg, 500)
		return
	}
	if len(chirp.Body) > 140 {
		responseWithJsonError(w, "Chirp is too long", 400)
		return
	}
	responseWithJson(struct {
		CleanedBody string `json:"cleaned_body"`
	}{CleanedBody: cleanProfaneWords(chirp.Body, []string{"kerfuffle", "sharbert", "fornax"})}, w, 200)
}

func (cfg *apiConfig) handleGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		responseWithJsonError(w, "Invalid chirp ID", 400)
		return
	}
	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			responseWithJsonError(w, "Chirp not found", 404)
		} else {
			responseWithJsonError(w, err.Error(), 500)
		}
		return
	}
	responseWithJson(mapChirpToResponse(&chirp), w, http.StatusOK)
}

func (cfg *apiConfig) handleGetChirps(w http.ResponseWriter, r *http.Request) {
	author_id := r.URL.Query().Get("author_id")
	sortOrder := r.URL.Query().Get("sort") // Get sorting order from query params
	// Default sorting order is "asc"
	sortOrder = strings.ToLower(sortOrder)
	if sortOrder != "desc" {
		sortOrder = "asc"
	}
	var chirps []database.Chirp
	var err error
	if author_id != "" {
		authorID, err := uuid.Parse(author_id)
		if err != nil {
			responseWithJsonError(w, "Invalid author ID", 400)
			return
		}
		chirps, err = cfg.db.GetChirpsByUser(r.Context(), database.GetChirpsByUserParams{
			UserID:  authorID,
			Column2: sortOrder,
		})
		if err != nil {
			responseWithJsonError(w, err.Error(), 500)
			return
		}
	} else {
		chirps, err = cfg.db.GetChirps(r.Context(), sortOrder)
		if err != nil {
			responseWithJsonError(w, err.Error(), 500)
			return
		}
	}
	var chirpsResponse []chirpResponse
	for _, chirp := range chirps {
		chirpsResponse = append(chirpsResponse, mapChirpToResponse(&chirp))
	}
	responseWithJson(chirpsResponse, w, http.StatusOK)
}

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request, _ string, user *database.User) {
	type jsonPayload struct {
		Body string `json:"body"`
	}
	jp := jsonPayload{}
	json.NewDecoder(r.Body).Decode(&jp)
	if jp.Body == "" {
		responseWithJsonError(w, "Body is required", 400)
		return
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   jp.Body,
		UserID: user.ID,
	})
	if err != nil {
		responseWithJsonError(w, err.Error(), 500)
		return
	}
	responseWithJson(mapChirpToResponse(&chirp), w, http.StatusCreated)
}

func (cfg *apiConfig) handleDeleteChirp(w http.ResponseWriter, r *http.Request, token string, user *database.User) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		responseWithJsonError(w, "Invalid chirp ID", 400)
		return
	}
	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		responseWithJsonError(w, err.Error(), 404)
		return
	}
	if chirp.UserID != user.ID {
		responseWithJsonError(w, "Forbidden", 403)
		return
	}
	err = cfg.db.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		responseWithJsonError(w, err.Error(), 404)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type chirpResponse struct {
	ID        uuid.UUID `json:"id"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func mapChirpToResponse(dc *database.Chirp) chirpResponse {
	return chirpResponse{
		ID:        dc.ID,
		Body:      dc.Body,
		UserID:    dc.UserID,
		CreatedAt: dc.CreatedAt,
		UpdatedAt: dc.UpdatedAt,
	}
}
