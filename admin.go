package main

import (
	"fmt"
	"html/template"
	"net/http"
)

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
	if cfg.platform != "dev" {
		w.WriteHeader(400)
		fmt.Fprintf(w, "This action requires administrator approval.  Nothing has changed.")
		return
	}
	err := cfg.db.DeleteUsers(req.Context())
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "An internal server error occurred while deleting all users: %s", err)
		return
	}
	w.WriteHeader(200)
	cfg.fileserverHits.Store(0)
	fmt.Fprintf(w, "OK")
}
