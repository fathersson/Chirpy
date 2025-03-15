package main

import (
	"Chirpy/domain"
	"Chirpy/internal/auth"
	"Chirpy/internal/database"
	"context"
	"net/http"

	"github.com/google/uuid"
)

func getUserId(cfg *apiConfig, header http.Header) (userId uuid.UUID, err error) {
	token, err := auth.GetBearerToken(header)
	if err != nil {
		return userId, err
	}
	userId, err = auth.ValidateJWT(token, cfg.tokenSecret)
	if err != nil {
		return userId, err
	}
	return userId, nil
}

func CreateRowUser(cfg *apiConfig, context context.Context, returnParams User) (user database.User, err error) {
	createHash(&returnParams)
	createUserParams := database.CreateUserParams{
		Email:          returnParams.Email,
		HashedPassword: returnParams.Password,
	}
	user, err = cfg.db.CreateUser(context, createUserParams)
	if err != nil {
		return database.User{}, err
	}
	return user, nil
}

func CreateRowChirp(cfg *apiConfig, context context.Context, cleanedChirp Chirp) (chirp database.Chirp, err error) {
	ChirpParams := database.CreateChirpParams{
		Body:   cleanedChirp.Body,
		UserID: cleanedChirp.UserID,
	}
	chirp, err = cfg.db.CreateChirp(context, ChirpParams)
	if err != nil {
		return database.Chirp{}, err
	}
	return chirp, nil
}

func GetChirpsAuthor(cfg *apiConfig, context context.Context, authorStr string) (chirpStruct []domain.Chirp, err error) {
	userId, err := uuid.Parse(authorStr)
	if err != nil {
		return []domain.Chirp{}, err
	}
	chirps, err := cfg.db.GetChirpsAuthorID(context, userId)
	if err != nil {
		return []domain.Chirp{}, err
	}
	chirpStruct = chirpsForStruct(chirps)
	return chirpStruct, nil
}

func getOneChirp(cfg *apiConfig, chirpId string, context context.Context) (chirp domain.Chirp, err error) {
	chirpParse, err := uuid.Parse(chirpId)
	if err != nil {
		return domain.Chirp{}, err
	}

	dbChirp, err := cfg.db.GetChirp(context, chirpParse)
	if err != nil {
		return domain.Chirp{}, err
	}
	chirpStruct := toChirp(dbChirp)
	return chirpStruct, nil
}
