package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/ZDSDD/Chirpy/internal/auth"
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
	chirpResponse, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		responseWithJsonError(w, err.Error(), 500)
		return
	}
	responseWithJson(mapChirpToResponse(&chirpResponse), w, http.StatusOK)
}

func (cfg *apiConfig) handleGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		responseWithJsonError(w, err.Error(), 500)
		return
	}
	var chirpsResponse []chirpResponse
	for _, chirp := range chirps {
		chirpsResponse = append(chirpsResponse, mapChirpToResponse(&chirp))
	}
	responseWithJson(chirpsResponse, w, http.StatusOK)
}

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	type jsonPayload struct {
		Body string `json:"body"`
	}
	jp := jsonPayload{}
	json.NewDecoder(r.Body).Decode(&jp)
	if jp.Body == "" {
		responseWithJsonError(w, "Body is required", 400)
		return
	}
	//Check JWT token
	tokenString, err := auth.GetBearerToken(r.Header)
	userId, err := auth.ValidateJWT(tokenString, cfg.jwtSecret)

	if err != nil {
		responseWithJsonError(w, err.Error(), 401)
		return
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   jp.Body,
		UserID: userId,
	})
	if err != nil {
		responseWithJsonError(w, err.Error(), 500)
		return
	}
	responseWithJson(mapChirpToResponse(&chirp), w, http.StatusCreated)
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
