package main

import (
	"Chirpy/internal/auth"
	"Chirpy/internal/database"
	"log"
	"net/http"
	"strings"
)

func checkApiKey(cfg *apiConfig, w http.ResponseWriter, r *http.Request) bool {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		cfg.respondWithError(w, http.StatusUnauthorized, "Invalid Authorization header")
		return false
	}
	if apiKey != cfg.polkaKey {
		cfg.respondWithError(w, http.StatusUnauthorized, "Invalid API key")
		return false
	}
	return true
}

func createHash(returnParams *User) {
	var err error
	returnParams.Password, err = auth.HashPassword(returnParams.Password)
	if err != nil {
		log.Println(err)
		return
	}
}

func toUser(dbUser database.User) User {
	return User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
		ChirpyRed: dbUser.IsChirpyRed.Bool,
	}
}

func toChirp(dbChirp database.Chirp) Chirp {
	return Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}
}

func cleanBody(body string) (bodyResp string) {
	bodySlice := strings.Split(body, " ")
	var bodySliceResp []string
	for _, v := range bodySlice {
		l := strings.ToLower(v)
		if l != "kerfuffle" && l != "sharbert" && l != "fornax" {
			bodySliceResp = append(bodySliceResp, v)
		} else {
			bodySliceResp = append(bodySliceResp, "****")
		}
	}
	bodyResp = strings.Join(bodySliceResp, " ")
	return
}

func chirpsForStruct(chirps []database.Chirp) (chirpStruct []Chirp) {
	for _, v := range chirps {
		chirpStruct = append(chirpStruct, toChirp(v))
	}
	return
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)

		next.ServeHTTP(w, r)
	})
}
