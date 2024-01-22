package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	apiCfg := apiConfig{
		fileserverHits: 0,
	}

	router := chi.NewRouter()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	router.Handle("/app", fsHandler)
	router.Handle("/app/*", fsHandler)

	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", readinessHandler)
	apiRouter.Get("/reset", apiCfg.resetHandler)
	apiRouter.Post("/validate_chirp", validateChirpHandler)
	router.Mount("/api", apiRouter)

	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", apiCfg.metricsHandler)
	router.Mount("/admin", adminRouter)

	corsMux := middlewareCors(router)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8 ")
	w.Write([]byte(http.StatusText(http.StatusOK)))
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

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var chirp struct {
		Body string `json:"body"`
	}
	if err := decoder.Decode(&chirp); err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error decoding parameters: %s", err))
		return
	}

	if err := validateChirpLength(w, chirp.Body, 140); err != nil {
		return
	}

	// params is a struct with data populated successfully

	validChirp := struct {
		CleanedBody string `json:"cleaned_body"`
	}{
		CleanedBody: cleanBody(chirp.Body),
	}

	err := respondWithJSON(w, 200, validChirp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error marshalling JSON: %s", err))
		return
	}
}

func validateChirpLength(w http.ResponseWriter, body string, maxLen int) error {
	if len(body) >= maxLen {
		chirpError := struct {
			Error string `json:"error"`
		}{
			Error: "Chirp is too long",
		}
		err := respondWithJSON(w, http.StatusBadRequest, chirpError)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return errors.New(fmt.Sprintf("Bad length, max length: %d, was: %d", maxLen, len(body)))
	}
	return nil
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	log.Print(msg)
	w.WriteHeader(code)
	w.Write([]byte(msg))
}
