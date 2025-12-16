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

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

func getRequestUserData(r *http.Request) (email string, hashedPw string, err error) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		return "", "", err
	}

	hashedPw, err = auth.HashPassword(params.Password)
	if err != nil {
		return "", "", err
	}

	return params.Email, hashedPw, nil
}

func (cfg *apiConfig) addUserHandler(w http.ResponseWriter, r *http.Request) {
	email, hashedPw, err := getRequestUserData(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not get params", err)
		return
	}
	userParams := database.CreateUserParams{
		Email:          email,
		HashedPassword: hashedPw,
	}

	dbUser, err := cfg.dbQueries.CreateUser(context.Background(), userParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating user", err)
		return
	}

	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}

	respondWithJSON(w, http.StatusCreated, user)
}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't decode parameters", err)
		return
	}

	dbUser, err := cfg.dbQueries.GetUserByEmail(context.Background(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't find user", err)
		return
	}

	isCorrectPW, err := auth.CheckPasswordHash(params.Password, dbUser.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	if isCorrectPW {
		expirationDuration := time.Duration(1) * time.Hour

		token, err := auth.MakeJWT(dbUser.ID, cfg.secret, expirationDuration)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't generate JWT", err)
			return
		}

		refreshToken, err := auth.MakeRefreshToken()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't generate refresh token", err)
			return
		}

		refreshTokenParams := database.CreateRefreshTokenParams{
			Token:  refreshToken,
			UserID: dbUser.ID,
		}
		_, err = cfg.dbQueries.CreateRefreshToken(context.Background(), refreshTokenParams)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't save refresh token", err)
			return
		}

		respondWithJSON(w, http.StatusOK, User{
			ID:           dbUser.ID,
			CreatedAt:    dbUser.CreatedAt,
			UpdatedAt:    dbUser.UpdatedAt,
			Email:        dbUser.Email,
			Token:        token,
			RefreshToken: refreshToken,
		})
	} else {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", errors.New("Incorrect email or password"))
	}
}

func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No Refresh Token", errors.New("No Refresh Token"))
		return
	}

	refreshTokenDB, err := cfg.dbQueries.GetRefreshToken(context.Background(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token not found", errors.New("Token not found"))
		return
	}

	if refreshTokenDB.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Token revoked", errors.New("Token revoked"))
		return
	}

	if refreshTokenDB.ExpiresAt.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "Token expired", errors.New("Token expired"))
		return
	}

	type refreshTokenRes struct {
		Token string `json:"token"`
	}
	newToken, err := auth.MakeJWT(refreshTokenDB.UserID, cfg.secret, time.Duration(1)*time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "New Token could not be generated", err)
		return
	}
	respondWithJSON(w, http.StatusOK, refreshTokenRes{Token: newToken})
}

func (cfg *apiConfig) revokeHandler(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No Refresh Token", errors.New("No Refresh Token"))
		return
	}

	err = cfg.dbQueries.RevokeToken(context.Background(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Refresh token not found", errors.New("Refresh token not found"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token not found", err)
		return
	}

	uuid, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	email, hashedPw, err := getRequestUserData(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not get params", err)
		return
	}
	userParams := database.UpdateUserParams{
		Email:          email,
		HashedPassword: hashedPw,
		ID:             uuid,
	}

	dbUser, err := cfg.dbQueries.UpdateUser(context.Background(), userParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating user", err)
		return
	}

	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}
	respondWithJSON(w, http.StatusOK, user)
}
