package main

import (
	"fmt"
	"net/http"
)

type apiConfig struct {
	hits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.hits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.hits)))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.hits = 0
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func healthHanlder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	fmt.Println("Starting server")

	config := apiConfig{}

	mux := http.NewServeMux()
	mux.Handle(
		"/app/",
		http.StripPrefix(
			"/app/",
			config.middlewareMetricsInc(http.FileServer(http.Dir("."))),
		),
	)
	mux.HandleFunc("/healthz", healthHanlder)
	mux.HandleFunc("/metrics", config.metricsHandler)
	mux.HandleFunc("/reset", config.resetHandler)

	corsMux := middlewareCors(mux)

	server := http.Server{
		Addr: "0.0.0.0:8080",
		Handler: corsMux,
	}

	err := server.ListenAndServe()

	fmt.Println("Server stopped")
	fmt.Println(err)
}
