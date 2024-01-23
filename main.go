package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"

	"github.com/ZDSDD/Chirpy/internal/database"
	"github.com/go-chi/chi/v5"
)

const (
	databasePath = "internal/database/database.json"
)

type DBconfig struct {
	localDB *database.DB
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	apiCfg := apiConfig{
		fileserverHits: 0,
	}
	DBconf := DBconfig{}

	if db, err := database.NewDB(databasePath); err == nil {
		DBconf.localDB = db
	} else {
		log.Fatal(err)
		return
	}

	router := chi.NewRouter()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	router.Handle("/app", fsHandler)
	router.Handle("/app/*", fsHandler)

	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", readinessHandler)
	apiRouter.Get("/reset", apiCfg.resetHandler)
	apiRouter.Post("/chirps", DBconf.postChirpHandler)
	apiRouter.Get("/chirps", DBconf.getChirpHandler)
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

func (dbConf *DBconfig) postChirpHandler(w http.ResponseWriter, r *http.Request) {
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
	newChirp, err := dbConf.localDB.CreateChirp(validatedChirp.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	respondWithJSON(w, 201, newChirp)
}

func (dbConf *DBconfig) getChirpHandler(w http.ResponseWriter, r *http.Request) {
	chirps, err := dbConf.localDB.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sort.Slice(chirps, func(i, j int) bool { return chirps[i].ID < chirps[j].ID })
	respondWithJSON(w, 200, chirps)
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8 ")
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
