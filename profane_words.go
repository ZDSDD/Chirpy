package main

import "strings"

var ProfaneWords = []string{"kerfuffle", "sharbert", "fornax"}

func cleanBody(chirp string) string{
	censored := "****"
	words := strings.Split(chirp, " ")
	for i, word := range words{
		for _,profaneWord := range ProfaneWords{
			if strings.ToLower(word) == profaneWord{
				words[i] = censored
			}
		}
	}
	return strings.Join(words," ")
}
