package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)


func cleanChirp(chirp string) string {
	keywords := []string{"kerfuffle", "sharbert", "fornax"}
	
	subwords := strings.Split(chirp, " ")
	for i, word := range subwords {
		lowerWord := strings.ToLower(word)
		for _, keyword := range keywords {
			if lowerWord == keyword {
				subwords[i] = "****"
			}
		}
	}
	result := strings.Join(subwords, " ")

	return result
}

type chirp struct {
	Chirp string `json:"body"`
}

func decodeChirp(r *http.Request) (string, error) {
	var c chirp
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&c)

	if err != nil {
		return "", err
	}

	if c.Chirp == "" {
		return "", errors.New("Chirp is empty")
	}

	return c.Chirp, nil
}
