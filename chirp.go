package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/TheMaru/go-http-server/internal/auth"
	"github.com/TheMaru/go-http-server/internal/database"
	"github.com/google/uuid"
)

type chirpResp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, r *http.Request) {
	const maxChirpLength = 140

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Not logged in", err)
	}

	uuid, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
	}
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	chirpParams := database.CreateChirpParams{
		Body:   filterProfanity(params.Body),
		UserID: uuid,
	}

	chirp, err := cfg.dbQueries.CreateChirp(context.Background(), chirpParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Chirp could not be created", err)
		return
	}

	chirpRes := chirpResp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	respondWithJSON(w, http.StatusCreated, chirpRes)
}

func (cfg *apiConfig) getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	chirpsFromDB, err := cfg.dbQueries.GetChirpsAsc(context.Background())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Chirps could not be loaded", err)
		return
	}

	chirpsResponse := make([]chirpResp, len(chirpsFromDB))
	for i, chirp := range chirpsFromDB {
		chirpsResponse[i] = chirpResp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
	}

	respondWithJSON(w, http.StatusOK, chirpsResponse)
}

func (cfg *apiConfig) getChirpByIDHandler(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "not a valid uuid", err)
	}

	chirpFromDB, err := cfg.dbQueries.GetChirpByID(context.Background(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
	}

	chirp := chirpResp{
		ID:        chirpFromDB.ID,
		CreatedAt: chirpFromDB.CreatedAt,
		UpdatedAt: chirpFromDB.UpdatedAt,
		Body:      chirpFromDB.Body,
		UserID:    chirpFromDB.UserID,
	}

	respondWithJSON(w, http.StatusOK, chirp)
}
