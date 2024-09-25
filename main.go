package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/ZDSDD/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(rw, req)
	})
}

func (cfg *apiConfig) handleReset(rw http.ResponseWriter, _ *http.Request) {
	cfg.fileserverHits.Store(0)
	rw.WriteHeader(http.StatusOK)
}

func (cfg *apiConfig) handleMetrics(rw http.ResponseWriter, _ *http.Request) {
	rw.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())))
}

// use godot package to load/read the .env file and
// return the value of the key
func getEnvVariable(key string) string {

	// load .env file
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

	log.Printf("Server run succesffuly on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
}

func (cfg *apiConfig) handleAdminMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())))
}

type Chirp struct {
	Body string `json:"body"`
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
	}{CleanedBody: cleanProfaneWords(chirp.Body, []string{"kerfuffle", "sharbert", "fornax"})}, w)
}

func responseWithJson(data interface{}, w http.ResponseWriter) {
	dat, ok := marshalToJson(data)
	if !ok {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
}

func marshalToJson(data interface{}) (dat []byte, ok bool) {
	dat, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return nil, false
	}
	return dat, true
}

func responseWithJsonError(w http.ResponseWriter, message string, errorCode int) {
	dat, ok := marshalToJson(struct {
		Error string `json:"error"`
	}{Error: message})
	if !ok {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorCode)
	w.Write(dat)
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
