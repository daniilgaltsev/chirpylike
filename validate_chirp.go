package main

import (
	"encoding/json"
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


func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Chirp *string `json:"body"`
	}

	type responseValid struct {
		CleanedBody string `json:"cleaned_body"`
	}

	type responseError struct {
		Error string `json:"error"`
	}


	var params parameters
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)

	if err != nil || params.Chirp == nil {
		response := responseError{Error: "Something went wrong"}
		dat, err := json.Marshal(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write(dat)
		return
	}

	if len(*params.Chirp) > 140 {
		response := responseError{Error: "Chirp is too long"}
		dat, err := json.Marshal(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write(dat)
		return
	}

	response := responseValid{CleanedBody: cleanChirp(*params.Chirp)}
	dat, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(dat)

}