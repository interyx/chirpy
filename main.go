package main

import _ "github.com/lib/pq"
import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
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
	muxer := http.NewServeMux()
	apiCfg := apiConfig{}
	muxer.Handle("/app/", apiCfg.middlewareMetricsInc(fileHandler()))
	muxer.HandleFunc("GET /api/healthz", readyHandler)
	muxer.HandleFunc("GET /admin/metrics", apiCfg.writeCountHandler)
	muxer.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	muxer.HandleFunc("POST /api/validate_chirp", validateChirp)
	server := http.Server{
		Handler: muxer,
		Addr:    ":8080",
	}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("An error has occurred: %v", err)
	}
}
