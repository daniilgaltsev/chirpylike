package main

import (
	"fmt"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
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
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(
		`
		<html>
		<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
		</body>
		</html>
		`,
		cfg.hits,
	)))
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


func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Chirp *string `json:"body"`
	}

	type responseValid struct {
		Valid bool `json:"valid"`
	}

	type responseError struct {
		Error string `json:"error"`
	}


	var params parameters
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&params)

	if err != nil || params.Chirp == nil {
		response := responseError{Error: "Something went wrong"}
		dat, err := json.Marshal(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write(dat)
		return
	}

	if len(*params.Chirp) > 140 {
		response := responseError{Error: "Chirp is too long"}
		dat, err := json.Marshal(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write(dat)
		return
	}

	response := responseValid{Valid: true}
	dat, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(dat)

}

func main() {
	fmt.Println("Starting server")

	config := apiConfig{}

	router := chi.NewRouter()

	fileServerHandler := config.middlewareMetricsInc(
		http.StripPrefix("/app", http.FileServer(http.Dir("."))),
	)
	router.Handle("/app/*", fileServerHandler)
	router.Handle("/app", fileServerHandler)
	
	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", healthHanlder)
	apiRouter.HandleFunc("/reset", config.resetHandler)
	apiRouter.Post("/validate_chirp", validateChirpHandler)
	router.Mount("/api", apiRouter)

	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", config.metricsHandler)
	router.Mount("/admin", adminRouter)

	corsRouter := middlewareCors(router)

	server := http.Server{
		Addr: "0.0.0.0:8080",
		Handler: corsRouter,
	}

	err := server.ListenAndServe()

	fmt.Println("Server stopped")
	fmt.Println(err)
}
