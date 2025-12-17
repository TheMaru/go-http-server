package main

import (
	"context"
	"encoding/json"
	"errors"
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
		return
	}

	uuid, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
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
	queryParamString := r.URL.Query().Get("author_id")

	var chirpsFromDB []database.Chirp
	var err error

	if queryParamString == "" {
		chirpsFromDB, err = cfg.dbQueries.GetChirpsAsc(context.Background())
	} else {
		userId, err := uuid.Parse(r.URL.Query().Get("author_id"))
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Author param malformed", err)
		}
		chirpsFromDB, err = cfg.dbQueries.GetChirpsByAuthor(context.Background(), userId)
	}
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
		respondWithError(w, http.StatusBadRequest, "Not a valid uuid", err)
		return
	}

	chirpFromDB, err := cfg.dbQueries.GetChirpByID(context.Background(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
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

func (cfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token not found", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	id, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Not a valid uuid", err)
		return
	}
	chirpFromDb, err := cfg.dbQueries.GetChirpByID(context.Background(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}

	if chirpFromDb.UserID != userID {
		respondWithError(w, http.StatusForbidden, "Not authorized to delete others chirps", errors.New("Forbidden"))
		return
	}

	err = cfg.dbQueries.DeleteChirp(context.Background(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Error in database query", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
