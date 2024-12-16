package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/interyx/chirpy/internal/auth"
	"github.com/interyx/chirpy/internal/database"
)

func readyHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		msg := fmt.Sprintf("An error has occurred: %v\n", err)
		respondWithError(w, 500, msg)
	}
}

func (cfg *apiConfig) addUser(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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
		msg := fmt.Sprintf("An error occurred marshaling JSON: %s", err)
		respondWithError(w, 400, msg)
		return
	}

	safePassword, err := auth.HashPassword(params.Password)

	if err != nil {
		msg := fmt.Sprintf("An error occurred generating a password: %s", err)
		respondWithError(w, 500, msg)
		return
	}

	userParameters := database.CreateUserParams{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Email:          params.Email,
		HashedPassword: safePassword,
	}
	user, err := cfg.db.CreateUser(req.Context(), userParameters)
	if err != nil {
		msg := fmt.Sprintf("An error occurred inserting the user into the database: %s", err)
		respondWithError(w, 500, msg)
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
		msg := fmt.Sprintf("An error occurred marshaling the JSON data: %s\n", err)
		respondWithError(w, 500, msg)
		return
	}
	respondWithJSON(w, 201, out)
}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Password         string `json:"password"`
		Email            string `json:"email"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}
	type outerface struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}
	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		msg := fmt.Sprintf("An error occurred marshaling JSON: %s", err)
		respondWithError(w, 400, msg)
		return
	}
	user, err := cfg.db.GetUserByEmail(req.Context(), params.Email)
	if err != nil {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}
	expirationTime := 60 * 60
	if params.ExpiresInSeconds > 0 {
		if params.ExpiresInSeconds < expirationTime {
			expirationTime = params.ExpiresInSeconds
		}
	}
	durationString := fmt.Sprintf("%vs", expirationTime)
	d, err := time.ParseDuration(durationString)
	if err != nil {
		respondWithError(w, 500, "Internal error parsing time")
	}
	token, err := auth.MakeJWT(user.ID, cfg.signJWT, d)
	if err != nil {
		respondWithError(w, 500, "Could not create JWT")
	}
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, 500, "Error generating refresh token")
	}
	data := outerface{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
	}
	out, err := json.Marshal(data)
	if err != nil {
		respondWithError(w, 500, "A marshaling error occurred")
		return
	}
	respondWithJSON(w, 200, out)
}
