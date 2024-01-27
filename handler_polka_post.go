package main

import (
	"encoding/json"
	"net/http"

	"github.com/ZDSDD/Chirpy/internal/auth"
)

func (cfg *apiConfig) handlerPolkaWebhooksPost(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID int `json:"user_id"`
		} `json:"data"`
	}

	apiKey, err := auth.GetToken(r.Header, "ApiKey")
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find apiKey")
		return
	}
	if apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "Bad polka api key")
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(200)
		return
	}

	err = cfg.DB.UpgradeUser(params.Data.UserID)
	if err != nil {
		respondWithError(w, 404, err.Error())
	}
	w.WriteHeader(200)
}
