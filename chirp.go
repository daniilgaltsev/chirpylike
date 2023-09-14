package main

import (
	"encoding/json"
	"net/http"

	"github.com/daniilgaltsev/chirpylike/internal/database"
)

func chirpsRespondWithInternalError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
}

func respondWithJson(w http.ResponseWriter, dat []byte, status int) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	w.Write(dat)
}


func chirpsRespondWithJsonError(w http.ResponseWriter, e string) {
	type responseError struct {
		Error string `json:"error"`
	}

	response := responseError{Error: e}
	dat, err := json.Marshal(response)
	if err != nil {
		chirpsRespondWithInternalError(w)
		return
	}

	respondWithJson(w, dat, http.StatusBadRequest)
}

func handleChirpsPost(w http.ResponseWriter, r *http.Request) {
	type responseValid struct {
		CleanedBody string `json:"cleaned_body"`
	}

	chirpBody, err := decodeChirp(r)

	if err != nil {
		chirpsRespondWithJsonError(w, "Something went wrong")
		return
	}

	if len(chirpBody) > 140 {
		chirpsRespondWithJsonError(w, "Chirp is too long")
		return
	}

	response := responseValid{CleanedBody: cleanChirp(chirpBody)}
	dat, err := json.Marshal(response)
	if err != nil {
		chirpsRespondWithInternalError(w)
		return
	}

	err = database.SaveChirp(chirpBody)
	if err != nil {
		chirpsRespondWithInternalError(w)
		return
	}
	
	respondWithJson(w, dat, http.StatusOK)
}