package main


import (
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"

	"github.com/daniilgaltsev/chirpylike/internal/database"
)

type user struct {
	Password string `json:"password"`
	Email string `json:"email"`
	Experies int `json:"expries_in_seconds"`
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

func createJWT(id, experiesInSeconds int, jwtSecret string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer: "chirpy",
		IssuedAt: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(
			now.Add(time.Duration(experiesInSeconds) * time.Second),
		),
		Subject: string(id),
	}
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)
	signedToken, err := token.SignedString([]byte(jwtSecret))
	return signedToken, err
}

func handleLoginPost(w http.ResponseWriter, r *http.Request, jwtSecret string) {
	type responseUser struct {
		Id int `json:"id"`
		Email string `json:"email"`
		Token string `json:"token"`
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

	experies := 60 * 60 * 24
	if u.Experies == 0 {
		u.Experies = experies
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

	token, err := createJWT(userToAuth.Id, u.Experies, jwtSecret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := responseUser{
		Id: userToAuth.Id,
		Email: userToAuth.Email,
		Token: token,
	}

	dat, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	respondWithJson(w, dat, http.StatusOK)
}
