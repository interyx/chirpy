package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/interyx/chirpy/internal/database"
	"log"
	"net/http"
	"strings"
	"time"
)

func readyHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		fmt.Printf("An error has occurred: %v\n", err)
	}
}

func (cfg *apiConfig) addUser(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}

	type returnVals struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}
	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding request body: %s\n", err)
		w.WriteHeader(400)
		return
	}
	userParameters := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email:     params.Email,
	}
	user, err := cfg.db.CreateUser(req.Context(), userParameters)
	if err != nil {
		fmt.Println("An error occurred inserting the user into the database: %s", err)
		w.WriteHeader(500)
		return
	}
	respBody := returnVals{
		Id:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	out, err := json.Marshal(respBody)
	if err != nil {
		fmt.Printf("An error occurred marshaling the JSON data: %s\n", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(201)
	w.Write(out)
}

func validateChirp(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type returnVals struct {
		CleanStr string `json:"cleaned_body"`
		Error    string `json:"error"`
	}
	w.Header().Set("Content-Type", "application/json")
	// decode request
	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	var respBody returnVals
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding request body: %s", err)
		w.WriteHeader(400)
		return
	}
	// construct response
	if len(params.Body) <= 140 {
		respBody.CleanStr = cleanString(params.Body)
		w.WriteHeader(200)
	} else {
		respBody.Error = "Chirp is too long"
		w.WriteHeader(400)
	}
	out, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling json: %s", err)
		w.WriteHeader(500)
		return
	}
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
