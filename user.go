package main


import (
	"encoding/json"
	"net/http"
	"strconv"

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

func handleUsersPut(w http.ResponseWriter, r *http.Request, jwtSecret string) {
	type responseUser struct {
		Id int `json:"id"`
		Email string `json:"email"`
	}

	_, claims, err := parseAuthorization(r.Header.Get("Authorization"), jwtSecret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	issuer, err := claims.GetIssuer()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if issuer != issuerAccess {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	stringId, err := claims.GetSubject()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, err := strconv.Atoi(stringId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var u user
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, err := database.UpdateUser(id, u.Email, u.Password)
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

	respondWithJson(w, dat, http.StatusOK)
}


func handleLoginPost(w http.ResponseWriter, r *http.Request, jwtSecret string) {
	type responseUser struct {
		Id int `json:"id"`
		Email string `json:"email"`
		Token string `json:"token"`
		RefreshToken string `json:"refresh_token"`
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

	token, err := createJWT(userToAuth.Id, jwtSecret, false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	refreshToken, err := createJWT(userToAuth.Id, jwtSecret, true)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := responseUser{
		Id: userToAuth.Id,
		Email: userToAuth.Email,
		Token: token,
		RefreshToken: refreshToken,
	}

	dat, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	respondWithJson(w, dat, http.StatusOK)
}
