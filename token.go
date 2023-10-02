package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/daniilgaltsev/chirpylike/internal/database"
)


const issuerAccess = "chirpy-access"
const issuerRefresh = "chirpy-refresh"


func createJWT(id int, jwtSecret string, isRefresh bool) (string, error) {
	issuer := issuerAccess
	expiresInSeconds := 60 * 60
	if isRefresh {
		issuer = issuerRefresh
		expiresInSeconds *= 24 * 60
	}
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer: issuer,
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

func parseAuthorization(authorization, jwtSecret string) (string, jwt.RegisteredClaims, error) {
	if authorization == "" {
		return "", jwt.RegisteredClaims{}, errors.New("missing authorization header")
	}

	bearerLength := len("Bearer ")
	if len(authorization) < bearerLength {
		return "", jwt.RegisteredClaims{}, errors.New("invalid authorization header")
	}
	tokenString := authorization[bearerLength:]
	claims := jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		},
	)
	if err != nil || !token.Valid {
		return tokenString, jwt.RegisteredClaims{}, errors.New("invalid token")
	}

	return tokenString, claims, nil
}


func handleRefreshPost(w http.ResponseWriter, r *http.Request, jwtSecret string) {
	type responseRefresh struct {
		Token string `json:"token"`
	}

	tokenString, claims, err := parseAuthorization(r.Header.Get("Authorization"), jwtSecret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	issuer, err := claims.GetIssuer()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if issuer != issuerRefresh {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	exists, err := database.RevokedTokenExists(tokenString)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if exists {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	subject, err := claims.GetSubject()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	id, err := strconv.Atoi(subject)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := createJWT(id, jwtSecret, false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := responseRefresh{
		Token: token,
	}

	dat, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	respondWithJson(w, dat, http.StatusOK)
}

func handleRevokePost(w http.ResponseWriter, r *http.Request, jwtSecret string) {
	tokenString, _, err := parseAuthorization(r.Header.Get("Authorization"), jwtSecret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = database.AddRevokedToken(tokenString)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
