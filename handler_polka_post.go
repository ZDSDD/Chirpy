package main

import (
	"net/http"
)

func (cfg *apiConfig) handlerPolkaWebhooksPost(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID int `json:"user_id"`
		} `json:"data"`
	}
}
