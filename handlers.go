package main

import (
	"Chirpy/internal/auth"
	"Chirpy/internal/database"

	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) webhooksHandler(w http.ResponseWriter, r *http.Request) {
	var webhookData Webhook
	if !checkApiKey(cfg, w, r) {
		return
	}
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&webhookData)
	if err != nil {
		cfg.respondWithError(w, http.StatusBadRequest, "Error decoding webhook")
		return
	}
	if webhookData.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if webhookData.Data.UserID == uuid.Nil {
		cfg.respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	_, err = cfg.db.CheckUser(r.Context(), webhookData.Data.UserID)
	if err != nil {
		cfg.respondWithError(w, http.StatusNotFound, "User Not Found")
		return
	}

	_, err = cfg.db.UpdateChirpyRed(r.Context(), webhookData.Data.UserID)
	if err != nil {
		cfg.respondWithError(w, http.StatusInternalServerError, "Could not update Chirpy Red")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	userId := getUserId(cfg, w, r)
	if userId == uuid.Nil {
		return
	}

	chirpStruct := getOneChirp(cfg, w, r)
	if chirpStruct.ID == uuid.Nil {
		return
	}

	if userId == chirpStruct.UserID {
		err := cfg.db.DeleteChirp(r.Context(), chirpStruct.ID)
		if err != nil {
			cfg.respondWithError(w, http.StatusInternalServerError, "Failed to delete chirp")
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}
	cfg.respondWithError(w, http.StatusForbidden, "This chirp not your (no access)")
}

func (cfg *apiConfig) putUsersHandler(w http.ResponseWriter, r *http.Request) {
	userId := getUserId(cfg, w, r)
	body := decoding(cfg, w, r)
	var err error
	body.Password, err = auth.HashPassword(body.Password)
	if err != nil {
		cfg.respondWithError(w, 500, "Error hashing password")
		return
	}
	updateUser, err := cfg.db.UpdateEMailPasword(r.Context(), database.UpdateEMailPaswordParams{
		Email:          body.Email,
		HashedPassword: body.Password,
		ID:             userId,
	})
	if err != nil {
		cfg.respondWithError(w, 500, "Error updating user")
		return
	}
	structUser := toUser(updateUser)
	cfg.respondWithJSON(w, 200, structUser)
}

func (cfg *apiConfig) revokeHandler(w http.ResponseWriter, r *http.Request) {
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		cfg.respondWithError(w, 401, "Error getting token")
		return
	}

	err = cfg.db.RevokeToken(r.Context(), bearerToken)
	if err != nil {
		cfg.respondWithError(w, 404, "Error revoking token")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request) {
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		cfg.respondWithError(w, 401, "Error getting token")
		return
	}
	RefreshToken, err := cfg.db.GetUserFromRefreshToken(r.Context(), bearerToken)
	if err != nil {
		cfg.respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if RefreshToken.ExpiresAt.Before(time.Now()) {
		cfg.respondWithError(w, http.StatusUnauthorized, "Refresh token expired")
		return
	}
	if RefreshToken.RevokedAt.Valid {
		cfg.respondWithError(w, http.StatusUnauthorized, "Refresh token revoked")
		return
	}

	newToken, err := auth.MakeJWT(RefreshToken.UserID, cfg.tokenSecret)
	if err != nil {
		cfg.respondWithError(w, http.StatusInternalServerError, "Error creating token")
		return
	}
	response := struct {
		Token string `json:"token"`
	}{
		Token: newToken,
	}
	cfg.respondWithJSON(w, http.StatusOK, response)
}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	returnParams := decoding(cfg, w, r)

	user, err := cfg.db.GetUser(r.Context(), returnParams.Email)
	if err != nil {
		cfg.respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}
	errCheck := auth.CheckPasswordHash(returnParams.Password, user.HashedPassword)
	if errCheck != nil {
		cfg.respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.tokenSecret)
	if err != nil {
		cfg.respondWithError(w, http.StatusInternalServerError, "Error creating token")
		return
	}
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		cfg.respondWithError(w, http.StatusInternalServerError, "Erorr creating refresh token")
		return
	}
	expiresAt := time.Now().Add(60 * 24 * time.Hour) // 60 days
	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		cfg.respondWithError(w, http.StatusInternalServerError, "Error storing refresh token")
		return
	}

	userStruct := toUser(user)
	ResponseStruct := LoginResponse{
		ID:           userStruct.ID,
		CreatedAt:    userStruct.CreatedAt,
		UpdatedAt:    userStruct.UpdatedAt,
		Email:        userStruct.Email,
		Token:        token,
		RefreshToken: refreshToken,
		ChirpyRed:    userStruct.ChirpyRed,
	}
	cfg.respondWithJSON(w, http.StatusOK, ResponseStruct)
}

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	returnParams := decoding(cfg, w, r)
	user := CreateRowUser(cfg, w, r, returnParams)
	userStruct := toUser(user)
	cfg.respondWithJSON(w, http.StatusCreated, userStruct)
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	cfg.db.DeleteUsers(r.Context())
	//он удалял всех пользователей из базы данных (не изменяя схему)
}

func (cfg *apiConfig) chirpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	UserID := getUserId(cfg, w, r)

	errLong := errStruct{
		Error: "Chirp is too long",
	}
	params := decodingChirp(cfg, w, r)
	if len(params.Body) > 140 {
		cfg.respondWithError(w, http.StatusBadRequest, errLong.Error)
		return
	}
	cleanedChirp := Chirp{
		Body:   cleanBody(params.Body),
		UserID: UserID,
	}

	chirp := CreateRowChirp(cfg, w, r, cleanedChirp)
	chirpStruct := toChirp(chirp)
	cfg.respondWithJSON(w, http.StatusCreated, chirpStruct)
}

func (cfg *apiConfig) getOneChirpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	chirpStruct := getOneChirp(cfg, w, r)
	cfg.respondWithJSON(w, http.StatusOK, chirpStruct)

}

func (cfg *apiConfig) getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sortStr := r.URL.Query().Get("sort")
	authorStr := r.URL.Query().Get("author_id")
	if authorStr != "" {
		GetChirpsAuthor(cfg, w, r, authorStr)
		return
	}

	chirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		cfg.respondWithError(w, http.StatusInternalServerError, "Something went wrong3")
		return
	}

	if sortStr == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		})
	} else {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
		})
	}

	chirpStruct := chirpsForStruct(chirps)
	cfg.respondWithJSON(w, http.StatusOK, chirpStruct)
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	bb := []byte(fmt.Sprintf(`
	<html>
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
	</html>`, cfg.fileserverHits.Load()))
	w.Write(bb)
}
