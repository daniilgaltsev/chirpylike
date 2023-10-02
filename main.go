package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	"github.com/daniilgaltsev/chirpylike/internal/database"
)

type apiConfig struct {
	hits int
	jwtSecret string
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

func (cfg *apiConfig) handleLoginPost(w http.ResponseWriter, r *http.Request) {
	handleLoginPost(w, r, cfg.jwtSecret)
}

func (cfg *apiConfig) handleUsersPut(w http.ResponseWriter, r *http.Request) {
	handleUsersPut(w, r, cfg.jwtSecret)
}

func (cfg *apiConfig) handleRefreshPost(w http.ResponseWriter, r *http.Request) {
	handleRefreshPost(w, r, cfg.jwtSecret)
}

func (cfg *apiConfig) handleRevokePost(w http.ResponseWriter, r *http.Request) {
	handleRevokePost(w, r, cfg.jwtSecret)
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

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		os.Exit(1)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		fmt.Println("JWT_SECRET is not set")
		os.Exit(1)
	}

	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	
	if *dbg {
		err = os.Remove(database.DbPath)
		if err != nil {
			fmt.Println(err)
		}
	}



	config := apiConfig{
		jwtSecret: jwtSecret,
	}

	router := chi.NewRouter()

	fileServerHandler := config.middlewareMetricsInc(
		http.StripPrefix("/app", http.FileServer(http.Dir("."))),
	)
	router.Handle("/app/*", fileServerHandler)
	router.Handle("/app", fileServerHandler)
	
	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", healthHanlder)
	apiRouter.HandleFunc("/reset", config.resetHandler)
	apiRouter.Post("/chirps", handleChirpsPost)
	apiRouter.Get("/chirps", handleChirpsGet)
	apiRouter.Get("/chirps/{id}", handleChirpsGetId)
	apiRouter.Post("/users", handleUsersPost)
	apiRouter.Put("/users", config.handleUsersPut)
	apiRouter.Post("/login", config.handleLoginPost)
	apiRouter.Post("/refresh", config.handleRefreshPost)
	apiRouter.Post("/revoke", config.handleRevokePost)
	router.Mount("/api", apiRouter)

	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", config.metricsHandler)
	router.Mount("/admin", adminRouter)

	corsRouter := middlewareCors(router)

	server := http.Server{
		Addr: "0.0.0.0:8080",
		Handler: corsRouter,
	}

	err = server.ListenAndServe()

	fmt.Println("Server stopped")
	fmt.Println(err)
}
