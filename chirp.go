package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/daniilgaltsev/chirpylike/internal/database"
)

func chirpsRespondWithInternalError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
}

func chirpsRespondWithBadRequestError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
}

func chirpsRespondWithNotFoundError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
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

func handleChirpsPost(w http.ResponseWriter, r *http.Request, jwtSecret string) {
	_, claims, err := parseAuthorization(r.Header.Get("Authorization"), jwtSecret)
	if err != nil {
		chirpsRespondWithBadRequestError(w)
		return
	}

	issuer, err := claims.GetIssuer()
	if err != nil {
		chirpsRespondWithInternalError(w)
		return
	}

	if issuer != issuerAccess {
		chirpsRespondWithBadRequestError(w)
		return
	}

	subject, err := claims.GetSubject()
	if err != nil {
		chirpsRespondWithInternalError(w) // TODO: definitely should fix all the errors to be consistent
		return
	}
	id, err := strconv.Atoi(subject)
	if err != nil {
		chirpsRespondWithInternalError(w)
		return
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

	chirp, err := database.SaveChirp(chirpBody, id)
	if err != nil {
		chirpsRespondWithInternalError(w)
		return
	}

	dat, err := json.Marshal(chirp)
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

func handleChirpsGetId(w http.ResponseWriter, r *http.Request) {
	db, err := database.GetDB()

	if err != nil {
		chirpsRespondWithInternalError(w)
		return
	}

	strId := chi.URLParam(r, "id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		chirpsRespondWithBadRequestError(w)
		return
	}
	
	chirp, ok := db.Chirps[id]
	if !ok {
		chirpsRespondWithNotFoundError(w)
		return
	}

	dat, err := json.Marshal(chirp)
	if err != nil {
		chirpsRespondWithInternalError(w)
		return
	}

	respondWithJson(w, dat, http.StatusOK)
}
