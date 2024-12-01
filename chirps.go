package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/interyx/chirpy/internal/database"
	"log"
	"net/http"
	"strings"
	"time"
)

func (cfg *apiConfig) getChirpsHandler(w http.ResponseWriter, req *http.Request) {
	chirps, err := cfg.db.GetChirps(req.Context())
	if err != nil {
		log.Printf("An error occurred getting values from the database: %s", err)
		w.WriteHeader(500)
		return
	}
	out, err := json.Marshal(chirps)
	if err != nil {
		log.Printf("An error occurred marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Write(out)
}

func (cfg *apiConfig) getChirpHandler(w http.ResponseWriter, req *http.Request) {
	id, err := uuid.Parse(req.PathValue("id"))
	if err != nil {
		log.Printf("UUID could not be parsed\n%s", err)
		w.WriteHeader(400)
		return
	}
	chirp, err := cfg.db.GetChirp(req.Context(), id)
	if err != nil {
		log.Printf("An error occurred retrieving the record: %s", err)
		w.WriteHeader(400)
		return
	}
	out, err := json.Marshal(chirp)
	if err != nil {
		log.Printf("An error occurred marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Write(out)
}

func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding request string: %s", err)
		w.WriteHeader(400)
		return
	}
	var chirp string
	if len(params.Body) <= 140 {
		chirp = cleanString(params.Body)
	} else {
		w.WriteHeader(400)
		return
	}
	chirpParams := database.CreateChirpParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body:      chirp,
		UserID:    params.UserID,
	}
	newChirp, err := cfg.db.CreateChirp(req.Context(), chirpParams)
	if err != nil {
		log.Printf("An error occurred while creating the chirp: %s\n", err)
		w.WriteHeader(400)
		return
	}
	out, err := json.Marshal(newChirp)
	if err != nil {
		log.Printf("An error occurred marshaling JSON data: %s", err)
		return
	}
	w.WriteHeader(201)
	w.Write(out)
}

func cleanString(str string) string {
	banned_words := []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}
	results := make([]string, len(str))
	wordsList := strings.Split(str, " ")
	for _, word := range wordsList {
		added := false
		for _, badWord := range banned_words {
			if strings.ToLower(word) == badWord {
				results = append(results, "****")
				added = true
				break
			}
		}
		if !added {
			results = append(results, word)
		}
	}
	return strings.Trim(strings.Join(results, " "), " ")
}
