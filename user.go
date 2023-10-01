package main


import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/daniilgaltsev/chirpylike/internal/database"
)

type user struct {
	Password string `json:"password"`
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
	if u.Email == "" || u.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := database.SaveUser(u.Email, u.Password)
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

func handleLoginPost(w http.ResponseWriter, r *http.Request) {
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
	if u.Email == "" || u.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db, err := database.GetDB()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ok := false
	var userToAuth database.User
	for _, userFromDb := range db.Users {
		if userFromDb.Email == u.Email {
			ok = true
			userToAuth = userFromDb
			break
		}
	}

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	hashedPassword := userToAuth.Password

	equal := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(u.Password))
	if equal != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	response := responseUser{
		Id: userToAuth.Id,
		Email: userToAuth.Email,
	}

	dat, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	respondWithJson(w, dat, http.StatusOK)
}
