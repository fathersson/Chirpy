package handler

import (
	"Chirpy/config"
	"Chirpy/decoding"
	"Chirpy/domain"
	"Chirpy/helpers"
	"Chirpy/internal/auth"
	"Chirpy/internal/database"
	"Chirpy/internal/service"
	"Chirpy/respond"
	"Chirpy/tokens"
	"fmt"
	"sort"
	"time"

	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// Handler структура для обработки HTTP запросов
type Handler struct {
	service     *service.Service
	serviceUser *service.ServiceUser
	cfg         *config.Config
}

// Конструктор для создания нового обработчика
func NewHandler(s *service.Service, ServiceUser *service.ServiceUser) *Handler {
	return &Handler{service: s, serviceUser: ServiceUser}
}

func (h Handler) WebhooksHandler(w http.ResponseWriter, r *http.Request) {
	var webhookData domain.Webhook
	if !helpers.CheckApiKey(h.cfg, w, r) {
		return
	}
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&webhookData)
	if err != nil {
		respond.RespondWithError(w, http.StatusBadRequest, "Error decoding webhook")
		return
	}
	if webhookData.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if webhookData.Data.UserID == uuid.Nil {
		respond.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	_, err = h.service.Db.CheckUser(r.Context(), webhookData.Data.UserID)
	if err != nil {
		respond.RespondWithError(w, http.StatusNotFound, "User Not Found")
		return
	}

	_, err = h.service.Db.UpdateChirpyRed(r.Context(), webhookData.Data.UserID)
	if err != nil {
		respond.RespondWithError(w, http.StatusInternalServerError, "Could not update Chirpy Red")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) DeleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := h.service.GetUserId(h.cfg, r.Header)
	if err != nil {
		respond.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	chirpId := r.PathValue("chirpID")
	context := r.Context()
	chirpStruct, err := h.serviceUser.DbU.GetOneChirp(chirpId, context)
	if err != nil {
		respond.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
	}
	if chirpStruct.ID == uuid.Nil {
		return
	}

	if userId != chirpStruct.UserID {
		respond.RespondWithError(w, http.StatusForbidden, "This chirp not your (no access)")
		return
	}

	err = h.service.Db.DeleteChirp(r.Context(), chirpStruct.ID)
	if err != nil {
		respond.RespondWithError(w, http.StatusInternalServerError, "Failed to delete chirp")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) PutUsersHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := h.serviceUser.DbU.GetUserId(h.cfg, r.Header)
	if err != nil {
		respond.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	body := decoding.Decoding(h.cfg, w, r)
	//var err error
	body.Password, err = auth.HashPassword(body.Password)
	if err != nil {
		respond.RespondWithError(w, 500, "Error hashing password")
		return
	}
	updateUser, err := h.service.Db.UpdateEMailPasword(r.Context(), database.UpdateEMailPaswordParams{
		Email:          body.Email,
		HashedPassword: body.Password,
		ID:             userId,
	})
	if err != nil {
		respond.RespondWithError(w, 500, "Error updating user")
		return
	}
	structUser := service.ToUser(updateUser)
	respond.RespondWithJSON(w, 200, structUser)
}

func (h Handler) RevokeHandler(w http.ResponseWriter, r *http.Request) {
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respond.RespondWithError(w, 401, "Error getting token")
		return
	}

	err = h.service.Db.RevokeToken(r.Context(), bearerToken)
	if err != nil {
		respond.RespondWithError(w, 404, "Error revoking token")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respond.RespondWithError(w, 401, "Error getting token")
		return
	}
	RefreshToken, err := h.service.Db.GetUserFromRefreshToken(r.Context(), bearerToken)
	if err != nil {
		respond.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if RefreshToken.ExpiresAt.Before(time.Now()) {
		respond.RespondWithError(w, http.StatusUnauthorized, "Refresh token expired")
		return
	}
	if RefreshToken.RevokedAt.Valid {
		respond.RespondWithError(w, http.StatusUnauthorized, "Refresh token revoked")
		return
	}

	newToken, err := auth.MakeJWT(RefreshToken.UserID, h.cfg.TokenSecret)
	if err != nil {
		respond.RespondWithError(w, http.StatusInternalServerError, "Error creating token")
		return
	}
	response := struct {
		Token string `json:"token"`
	}{
		Token: newToken,
	}
	respond.RespondWithJSON(w, http.StatusOK, response)
}

func (h Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	returnParams := decoding.Decoding(h.cfg, w, r)

	user, err := h.service.Db.GetUser(r.Context(), returnParams.Email)
	if err != nil {
		respond.RespondWithError(w, http.StatusUnauthorized, "Incorrect email")
		return
	}

	err = tokens.CheckPassword(returnParams.Password, user.HashedPassword)
	if err != nil {
		respond.RespondWithError(w, http.StatusUnauthorized, "Incorrect password")
		return
	}
	token, err := tokens.GetToken(user.ID, h.cfg.TokenSecret)
	if err != nil {
		respond.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%v", err))
		return
	}

	_, err = h.service.Db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     token.RefreshToken,
		UserID:    user.ID,
		ExpiresAt: token.ExpiresAt,
	})
	if err != nil {
		respond.RespondWithError(w, http.StatusInternalServerError, "Error storing refresh token")
		return
	}

	userStruct := service.ToUser(user)
	ResponseStruct := domain.LoginResponse{
		ID:           userStruct.ID,
		CreatedAt:    userStruct.CreatedAt,
		UpdatedAt:    userStruct.UpdatedAt,
		Email:        userStruct.Email,
		Token:        token.Token,
		RefreshToken: token.RefreshToken,
		ChirpyRed:    userStruct.ChirpyRed,
	}
	respond.RespondWithJSON(w, http.StatusOK, ResponseStruct)
}

func (h Handler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	returnParams := decoding.Decoding(h.cfg, w, r)
	context := r.Context()
	DbUser := service.ToDbUser(returnParams)
	user, err := h.serviceUser.DbU.CreateRowUser(context, DbUser)
	if err != nil {
		respond.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
	}
	respond.RespondWithJSON(w, http.StatusCreated, user)
}

func (h Handler) ResetHandler(w http.ResponseWriter, r *http.Request) {
	if h.cfg.Platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	h.cfg.FileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	h.service.Db.DeleteUsers(r.Context())
	//он удалял всех пользователей из базы данных (не изменяя схему)
}

func (h Handler) ChirpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userId, err := h.serviceUser.DbU.GetUserId(h.cfg, r.Header)
	if err != nil {
		respond.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	errLong := domain.ErrStruct{
		Error: "Chirp is too long",
	}
	params := decoding.DecodingChirp(h.cfg, w, r)
	if len(params.Body) > 140 {
		respond.RespondWithError(w, http.StatusBadRequest, errLong.Error)
		return
	}
	cleanedChirp := database.Chirp{
		Body:   helpers.CleanBody(params.Body),
		UserID: userId,
	}
	context := r.Context()
	chirp, err := h.serviceUser.DbU.CreateRowChirp(context, cleanedChirp)
	if err != nil {
		respond.RespondWithError(w, http.StatusInternalServerError, "Error creating chirp")
	}
	respond.RespondWithJSON(w, http.StatusCreated, chirp)
}

func (h Handler) GetOneChirpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	chirpId := r.PathValue("chirpID")
	context := r.Context()
	chirpStruct, err := h.serviceUser.DbU.GetOneChirp(chirpId, context)
	if err != nil {
		respond.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
	}
	respond.RespondWithJSON(w, http.StatusOK, chirpStruct)

}

func (h Handler) GetChirpsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sortStr := r.URL.Query().Get("sort")
	authorStr := r.URL.Query().Get("author_id")
	context := r.Context()
	if authorStr == "" {
		respond.RespondWithError(w, http.StatusBadRequest, "unknown author id")
		return
	}

	chirps, err := h.serviceUser.DbU.GetChirpsAuthor(context, authorStr, sortStr)
	if err != nil {
		respond.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
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

	respond.RespondWithJSON(w, http.StatusOK, chirps)
}

func (h Handler) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	bb := []byte(fmt.Sprintf(`
	<html>
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
	</html>`, h.cfg.FileserverHits.Load()))
	w.Write(bb)
}

func MiddlewareMetricsInc(cfg *config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits.Add(1)

		next.ServeHTTP(w, r)
	})
}
