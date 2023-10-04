package main

import (
	"encoding/json"
	"net/http"

	"github.com/daniilgaltsev/chirpylike/internal/database"
)

func parseAuthorizationPolkaApiKey(authorization, polkaApiKey string) bool {
	if authorization == "" {
		return false
	}

	apiKeyLength := len("ApiKey ")
	if len(authorization) < apiKeyLength {
		return false
	}
	apiKey := authorization[apiKeyLength:]
	
	if apiKey != polkaApiKey {
		return false
	}

	return true
}


func handlePolkaWebhooksPost(w http.ResponseWriter, r *http.Request, polkaApiKey string) {
	type requestBody struct {
		Event string `json:"event"`
		Data struct {
			UserId int `json:"user_id"`
		}
	}

	isValid := parseAuthorizationPolkaApiKey(r.Header.Get("Authorization"), polkaApiKey)
	if !isValid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var body requestBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if body.Event != "user.upgraded" {
		w.WriteHeader(http.StatusOK)
		return
	}

	err = database.UpdateUserMembership(body.Data.UserId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}
