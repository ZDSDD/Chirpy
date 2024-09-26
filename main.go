package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/ZDSDD/Chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func getEnvVariable(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv(key)
}

func main() {
	godotenv.Load()
	mux := http.NewServeMux()
	port := getEnvVariable("PORT")
	dbURL := getEnvVariable("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
		os.Exit(1)
	}
	dbQueries := database.New(db)

	cfg := &apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
	}

	server := http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}
	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", handleHealthz)
	mux.HandleFunc("POST /api/reset", cfg.handleReset)
	mux.HandleFunc("POST /admin/reset", cfg.handleReset)
	mux.HandleFunc("GET /api/metrics", cfg.handleMetrics)
	mux.HandleFunc("GET /admin/metrics", cfg.handleAdminMetrics)
	mux.HandleFunc("POST /api/validate_chirp", validateChirp)
	mux.HandleFunc("POST /api/users", cfg.handleCreateUser)
	mux.HandleFunc("POST /api/chirps", cfg.handleCreateChirp)
	mux.HandleFunc("GET /api/chirps", cfg.handleGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.handleGetChirp)
	log.Printf("Server run succesffuly on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
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

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
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
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}
	jp := jsonPayload{}
	json.NewDecoder(r.Body).Decode(&jp)
	if jp.Body == "" {
		responseWithJsonError(w, "Body is required", 400)
		return
	}
	if jp.UserId == uuid.Nil {
		responseWithJsonError(w, "User ID is required", 400)
	}
	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   jp.Body,
		UserID: jp.UserId,
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

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type UserReqBody struct {
		Email string `json:"email"`
	}

	var userReq UserReqBody
	json.NewDecoder(r.Body).Decode(&userReq)
	if userReq.Email == "" {
		responseWithJsonError(w, "Email is required", 400)
		return
	}
	user, err := cfg.db.CreateUser(r.Context(), userReq.Email)
	if err != nil {
		responseWithJsonError(w, err.Error(), 500)
		return
	}
	responseWithJson(mapUserToResponse(&user), w, http.StatusCreated)
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func mapUserToResponse(du *database.User) UserResponse {
	return UserResponse{
		ID:        du.ID,
		Email:     du.Email,
		CreatedAt: du.CreatedAt,
		UpdatedAt: du.UpdatedAt,
	}
}
