package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/TheMaru/go-http-server/internal/auth"
	"github.com/google/uuid"
)

type polkaRequest struct {
	Event string `json:"event"`
	Data  struct {
		UserID uuid.UUID `json:"user_id"`
	} `json:"data"`
}

func (cfg *apiConfig) polkaWebhookHandler(w http.ResponseWriter, r *http.Request) {
	key, err := auth.GetAPIKey(r.Header)
	if err != nil || key != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "Not authorized", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	requestParams := polkaRequest{}
	err = decoder.Decode(&requestParams)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couln't decode parameters", err)
		return
	}

	if requestParams.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err = cfg.dbQueries.GrantChirpyRedToUser(context.Background(), requestParams.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Could not update user", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
