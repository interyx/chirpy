package main

import (
	"fmt"
	"html/template"
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
	server := http.Server{
		Handler: muxer,
		Addr:    ":8080",
	}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("An error has occurred: %v", err)
	}
}

func readyHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		fmt.Printf("An error has occurred: %v\n", err)
	}
}

func (cfg *apiConfig) writeCountHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	tmpl, err := template.ParseFiles("hits.html")
	if err != nil {
		fmt.Printf("An error has occured: %v\n", err)
	}
	data := struct {
		Hits int32
	}{
		Hits: cfg.fileserverHits.Load(),
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		fmt.Printf("An error has occurred: %v\n", err)
	}
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	cfg.fileserverHits.Store(0)
	fmt.Fprintf(w, "OK")
}
