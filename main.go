package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/ZDSDD/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

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
		jwtSecret:      getEnvVariable("JWT_SECRET"),
	}

	server := http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}
	// Static file handling
	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	// Health check and Metrics endpoints
	mux.HandleFunc("GET /api/healthz", handleHealthz)
	mux.HandleFunc("GET /api/metrics", cfg.handleMetrics)
	mux.HandleFunc("GET /admin/metrics", cfg.handleAdminMetrics)

	// User-related routes
	mux.HandleFunc("POST /api/users", cfg.handleCreateUser)
	mux.HandleFunc("POST /api/login", cfg.handleLogin)
	mux.HandleFunc("PUT /api/users", cfg.requireBearerToken(cfg.handleUpdateUser))
	mux.HandleFunc("POST /api/polka/webhooks", cfg.handleUpgradePolkaUser)

	// JWT-related routers
	mux.HandleFunc("POST /api/refresh", cfg.requireBearerToken(cfg.requireValidJWTToken(cfg.handleRefreshToken)))
	mux.HandleFunc("POST /api/revoke", cfg.requireBearerToken(cfg.handleRevokeToken))

	// Chirps-related routes
	mux.HandleFunc("POST /api/chirps", cfg.requireBearerToken(cfg.requireValidJWTToken(cfg.handleCreateChirp)))
	mux.HandleFunc("GET /api/chirps", cfg.handleGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.handleGetChirp)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.requireBearerToken(cfg.requireValidJWTToken(cfg.handleDeleteChirp)))
	mux.HandleFunc("POST /api/validate_chirp", validateChirp)

	// Admin-related routes
	mux.HandleFunc("POST /admin/reset", cfg.handleReset)

	// Miscellaneous routes
	mux.HandleFunc("POST /api/reset", cfg.handleReset)

	// Start the server
	log.Printf("Server running successfully on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
}

func getEnvVariable(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv(key)
}
