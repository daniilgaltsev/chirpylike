package main

import (
	"encoding/json"
	"net/http"

	"github.com/daniilgaltsev/chirpylike/internal/database"
)


func handlePolkaWebhooksPost(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Event string `json:"event"`
		Data struct {
			UserId int `json:"user_id"`
		}
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
