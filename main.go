package main

import _ "github.com/lib/pq"
import (
	"database/sql"
	"fmt"
	"github.com/interyx/chirpy/internal/database"
	"github.com/joho/godotenv"
	"net/http"
	"os"
	"sync/atomic"
)

type apiConfig struct {
	db             *database.Queries
	fileserverHits atomic.Int32
	platform       string
	signJWT        string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func fileHandler() http.Handler {
	server := http.FileServer(http.Dir("."))
	return http.StripPrefix("/app/", server)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Cannot find .env.\nAn environment file with the database string is required.\n")
	}
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	tokenSecret := os.Getenv("SIGN_KEY")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("An error occurred opening the database: %s\n", err)
	}
	muxer := http.NewServeMux()
	dbQueries := database.New(db)
	apiCfg := apiConfig{
		db:       dbQueries,
		platform: platform,
		signJWT:  tokenSecret,
	}
	muxer.Handle("/app/", apiCfg.middlewareMetricsInc(fileHandler()))
	muxer.HandleFunc("GET /api/healthz", readyHandler)
	muxer.HandleFunc("GET /admin/metrics", apiCfg.writeCountHandler)
	muxer.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	muxer.HandleFunc("POST /api/chirps", apiCfg.createChirpHandler)
	muxer.HandleFunc("POST /api/users", apiCfg.addUser)
	muxer.HandleFunc("GET /api/chirps", apiCfg.getChirpsHandler)
	muxer.HandleFunc("GET /api/chirps/{id}", apiCfg.getChirpHandler)
	muxer.HandleFunc("POST /api/login", apiCfg.loginHandler)
	server := http.Server{
		Handler: muxer,
		Addr:    ":8080",
	}
	err = server.ListenAndServe()
	if err != nil {
		fmt.Printf("An error has occurred: %v\n", err)
	}
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "text/plain;encoding=utf-8")
	fmt.Fprintf(w, "%s", msg)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	out, ok := payload.([]byte)
	if !ok {
		respondWithError(w, 500, "An internal error occurred while writing the payload")
	}
	w.Write(out)
}
