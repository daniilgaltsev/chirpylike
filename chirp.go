package main

import (
	"encoding/json"
	"net/http"
	"sort"

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
	type responseChirp struct {
		Id int `json:"id"`
		Body string `json:"body"`
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

	chirpBody = cleanChirp(chirpBody)

	chirp, err := database.SaveChirp(chirpBody)
	if err != nil {
		chirpsRespondWithInternalError(w)
		return
	}

	response := responseChirp{
		Id: chirp.Id,
		Body: chirp.Body,
	}

	dat, err := json.Marshal(response)
	if err != nil {
		chirpsRespondWithInternalError(w)
		return
	}
	
	respondWithJson(w, dat, http.StatusCreated)
}

func handleChirpsGet(w http.ResponseWriter, r *http.Request) {
	db, err := database.GetDB()

	if err != nil {
		chirpsRespondWithInternalError(w)
		return
	}

	chirps := make([]database.Chirp, 0, len(db.Chirps))
	for _, chirp := range db.Chirps {
		chirps = append(chirps, chirp)
	}
	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].Id < chirps[j].Id
	})

	dat, err := json.Marshal(chirps)
	if err != nil {
		chirpsRespondWithInternalError(w)
		return
	}

	respondWithJson(w, dat, http.StatusOK)
}
