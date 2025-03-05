package main

import (
	"Chirpy/internal/auth"
	"Chirpy/internal/database"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

func getUserId(cfg *apiConfig, w http.ResponseWriter, r *http.Request) (userId uuid.UUID) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		cfg.respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userId, err = auth.ValidateJWT(token, cfg.tokenSecret)
	if err != nil {
		cfg.respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	return userId
}

func CreateRowUser(cfg *apiConfig, w http.ResponseWriter, r *http.Request, returnParams User) (user database.User) {
	createHash(&returnParams)
	createUserParams := database.CreateUserParams{
		Email:          returnParams.Email,
		HashedPassword: returnParams.Password,
	}
	user, err := cfg.db.CreateUser(r.Context(), createUserParams)
	if err != nil {
		fmt.Printf("Error creating user: %v\n", err) // Выводим ошибку для отладки
		cfg.respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	return user
}

func CreateRowChirp(cfg *apiConfig, w http.ResponseWriter, r *http.Request, cleanedChirp Chirp) (chirp database.Chirp) {
	ChirpParams := database.CreateChirpParams{
		Body:   cleanedChirp.Body,
		UserID: cleanedChirp.UserID,
	}
	chirp, err := cfg.db.CreateChirp(r.Context(), ChirpParams)
	if err != nil {
		fmt.Printf("Error creating chirp: %v\n", err) // Выводим ошибку для отладки
		cfg.respondWithError(w, http.StatusInternalServerError, "Something went wrong1")
		return
	}
	return chirp
}

func GetChirpsAuthor(cfg *apiConfig, w http.ResponseWriter, r *http.Request, authorStr string) {
	userId, err := uuid.Parse(authorStr)
	if err != nil {
		cfg.respondWithError(w, http.StatusBadRequest, "Invalid author ID")
		return
	}
	chirps, err := cfg.db.GetChirpsAuthorID(r.Context(), userId)
	if err != nil {
		cfg.respondWithError(w, http.StatusInternalServerError, "User not found")
		return
	}
	chirpStruct := chirpsForStruct(chirps)
	cfg.respondWithJSON(w, http.StatusOK, chirpStruct)
	return
}

func getOneChirp(cfg *apiConfig, w http.ResponseWriter, r *http.Request) (chirpStruct Chirp) {
	chirpID := r.PathValue("chirpID")
	// Преобразуем строку в uuid.UUID
	chirpParse, err := uuid.Parse(chirpID)
	if err != nil {
		cfg.respondWithError(w, http.StatusBadRequest, "Invalid chirpID format")
		return
	}
	chirp, err := cfg.db.GetChirp(r.Context(), chirpParse)
	if err != nil {
		cfg.respondWithError(w, http.StatusNotFound, "Chirp not found")
		return
	}
	chirpStruct = toChirp(chirp)
	return
}
