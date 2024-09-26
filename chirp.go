package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
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
