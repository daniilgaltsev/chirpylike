package main


import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"

	"github.com/daniilgaltsev/chirpylike/internal/database"
)

type user struct {
	Password string `json:"password"`
	Email string `json:"email"`
	Expires int `json:"expires_in_seconds"`
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

	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	tokenString := authorization[len("Bearer "):]
	claims := jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		},
	)
	if err != nil || !token.Valid {
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

func createJWT(id, expiresInSeconds int, jwtSecret string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer: "chirpy",
		IssuedAt: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(
			now.Add(time.Duration(expiresInSeconds) * time.Second),
		),
		Subject: strconv.Itoa(id),
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

	expires := 60 * 60 * 24
	if u.Expires == 0 {
		u.Expires = expires
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

	token, err := createJWT(userToAuth.Id, u.Expires, jwtSecret)
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
