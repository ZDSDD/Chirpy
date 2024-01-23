package main

import (
	"log"
	"net/http"

	"github.com/ZDSDD/Chirpy/internal/database"
	"github.com/go-chi/chi/v5"
)

const (
	databasePath = "internal/database/database.json"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	db, err := database.NewDB(databasePath)
	if err != nil {
		log.Fatal(err)
	}

	apiCfg := apiConfig{
		fileserverHits: 0,
		localDB:        db,
	}

	router := chi.NewRouter()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	router.Handle("/app", fsHandler)
	router.Handle("/app/*", fsHandler)

	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", readinessHandler)
	apiRouter.Get("/reset", apiCfg.resetHandler)
	apiRouter.Post("/chirps", apiCfg.postChirpHandler)
	apiRouter.Get("/chirps", apiCfg.getChirpHandler)
	apiRouter.Get("/chirps/{chirpID}", apiCfg.getChirpByIDHandler)
	apiRouter.Post("/users", apiCfg.postUserHandler)
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
