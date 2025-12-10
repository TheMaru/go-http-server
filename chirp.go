package main

import (
	"encoding/json"
	"net/http"
)

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	const maxChirpLength = 140
	type parameters struct {
		Body string `json:"body"`
	}
	type sucRes struct {
		CleanedBody string `json:"cleaned_body"`
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

	respondWithJSON(w, http.StatusOK, sucRes{CleanedBody: filterProfanity(params.Body)})
}
