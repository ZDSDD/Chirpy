package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

func validateChirpBody(chirpBody string, w http.ResponseWriter) error {
	if err := validateChirpLength(w, chirpBody, 140); err != nil {
		return err
	}
	return nil
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

var ProfaneWords = []string{"kerfuffle", "sharbert", "fornax"}

func cleanBody(chirp string) string {
	censored := "****"
	words := strings.Split(chirp, " ")
	for i, word := range words {
		for _, profaneWord := range ProfaneWords {
			if strings.ToLower(word) == profaneWord {
				words[i] = censored
			}
		}
	}
	return strings.Join(words, " ")
}
