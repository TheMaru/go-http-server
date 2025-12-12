package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

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

	type parameters struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
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
		UserID: params.UserId,
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
