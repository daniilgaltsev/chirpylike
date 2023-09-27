package main


import (
	"encoding/json"
	"net/http"

	"github.com/daniilgaltsev/chirpylike/internal/database"
)

type user struct {
	Email string `json:"email"`
}


func handleUsersPost(w http.ResponseWriter, r *http.Request) {
	type responseUser struct {
		Id int `json:"id"`
		Email string `json:"email"`
	}

	var u user
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&u)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if u.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := database.SaveUser(u.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := responseUser{
		Id: user.Id,
		Email: user.Email,
	}

	dat, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	respondWithJson(w, dat, http.StatusCreated)
}
