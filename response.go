package main


import (
	"net/http"
)


func respondWithJson(w http.ResponseWriter, dat []byte, status int) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	w.Write(dat)
}