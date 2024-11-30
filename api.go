package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func readyHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		fmt.Printf("An error has occurred: %v\n", err)
	}
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
