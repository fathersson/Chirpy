package handler

import (
	"Chirpy/internal/auth"
	"Chirpy/internal/config"
	"Chirpy/internal/database"
	"Chirpy/internal/decoding"
	"Chirpy/internal/domain"
	"Chirpy/internal/helpers"
	"Chirpy/internal/respond"
	"Chirpy/internal/service"

	//"Chirpy/internal/service"
	"Chirpy/internal/tokens"
	"context"
	"fmt"
	"sort"
	"time"

	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type ServiceInt interface {
	GetUserId(Cfg *config.Config, header http.Header) (userId uuid.UUID, err error)
	CreateRowUser(context context.Context, returnParams domain.User) (user domain.User, err error)
	CreateRowChirp(context context.Context, cleanedChirp domain.Chirp) (chirp domain.Chirp, err error)
	GetChirpsAuthor(context context.Context, authorStr string, sortOrder string) (chirpStruct []domain.Chirp, err error)
	GetOneChirp(chirpId string, context context.Context) (chirp domain.Chirp, err error)
}

/*type ServiceStr struct {
	Service *ServiceInt
}

func NewService(Service *ServiceInt) *ServiceStr {
	return &ServiceStr{Service: Service}
}*/

// Handler структура для обработки HTTP запросов
type Handler struct {
	db      service.Database
	service ServiceInt
	cfg     *config.Config
}

// Конструктор для создания нового обработчика
func NewHandler(db service.Database, service ServiceInt) *Handler {
	return &Handler{db: db, service: service}
}

func (h Handler) WebhooksHandler(w http.ResponseWriter, r *http.Request) {
	var webhookData domain.Webhook
	if !helpers.CheckApiKey(r.Header, h.cfg.POLKAKEY) {
		respond.RespondWithError(w, http.StatusUnauthorized, "Invalid Authorization header or invalid API key")
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

	_, err = h.db.CheckUser(r.Context(), webhookData.Data.UserID)
	if err != nil {
		respond.RespondWithError(w, http.StatusNotFound, "User Not Found")
		return
	}

	_, err = h.db.UpdateChirpyRed(r.Context(), webhookData.Data.UserID)
	if err != nil {
		respond.RespondWithError(w, http.StatusInternalServerError, "Could not update Chirpy Red")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) DeleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := GetUserId(h.cfg, r.Header)
	if err != nil {
		respond.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	chirpId := r.PathValue("chirpID")
	context := r.Context()
	chirpStruct, err := h.service.GetOneChirp(chirpId, context)
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

	err = h.db.DeleteChirp(r.Context(), chirpStruct.ID)
	if err != nil {
		respond.RespondWithError(w, http.StatusInternalServerError, "Failed to delete chirp")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) PutUsersHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := h.service.GetUserId(h.cfg, r.Header)
	if err != nil {
		respond.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Читаем тело запроса в байтовый срез
	body := make([]byte, r.ContentLength)
	_, err = r.Body.Read(body)
	if err != nil {
		respond.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	// Декодируем тело запроса в структуру User
	params, err := decoding.DecodingUser(body)
	if err != nil {
		respond.RespondWithError(w, http.StatusBadRequest, "Invalid user data")
		return
	}

	params.Password, err = auth.HashPassword(params.Password)
	if err != nil {
		respond.RespondWithError(w, 500, "Error hashing password")
		return
	}
	updateUser, err := h.db.UpdateEMailPasword(r.Context(), database.UpdateEMailPaswordParams{
		Email:          params.Email,
		HashedPassword: params.Password,
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

	err = h.db.RevokeToken(r.Context(), bearerToken)
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
	RefreshToken, err := h.db.GetUserFromRefreshToken(r.Context(), bearerToken)
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

	newToken, err := auth.MakeJWT(RefreshToken.UserID, h.cfg.TOKENSECRET)
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
	// Читаем тело запроса в байтовый срез
	body := make([]byte, r.ContentLength)
	_, err := r.Body.Read(body)
	if err != nil {
		respond.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	// Декодируем тело запроса в структуру User
	params, err := decoding.DecodingUser(body)
	if err != nil {
		respond.RespondWithError(w, http.StatusBadRequest, "Invalid user data")
		return
	}

	user, err := h.db.GetUser(r.Context(), params.Email)
	if err != nil {
		respond.RespondWithError(w, http.StatusUnauthorized, "Incorrect email")
		return
	}

	err = tokens.CheckPassword(params.Password, user.HashedPassword)
	if err != nil {
		respond.RespondWithError(w, http.StatusUnauthorized, "Incorrect password")
		return
	}
	token, err := tokens.GetToken(user.ID, h.cfg.TOKENSECRET)
	if err != nil {
		respond.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%v", err))
		return
	}

	_, err = h.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
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
	// Читаем тело запроса в байтовый срез
	body := make([]byte, r.ContentLength)
	_, err := r.Body.Read(body)
	if err != nil {
		respond.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	// Декодируем тело запроса в структуру User
	params, err := decoding.DecodingUser(body)
	if err != nil {
		respond.RespondWithError(w, http.StatusBadRequest, "Invalid user data")
		return
	}

	context := r.Context()
	DbUser := service.ToDbUser(params)
	user, err := h.service.CreateRowUser(context, DbUser)
	if err != nil {
		respond.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
	}
	respond.RespondWithJSON(w, http.StatusCreated, user)
}

func (h Handler) ResetHandler(w http.ResponseWriter, r *http.Request) {
	if h.cfg.PLATFORM != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	h.cfg.FILESERVERHITS.Store(0)
	w.WriteHeader(http.StatusOK)
	h.db.DeleteUsers(r.Context())
	//он удалял всех пользователей из базы данных (не изменяя схему)
}

func (h Handler) ChirpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userId, err := h.service.GetUserId(h.cfg, r.Header)
	if err != nil {
		respond.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	errLong := domain.ErrStruct{
		Error: "Chirp is too long",
	}

	// Читаем тело запроса в байтовый срез
	body := make([]byte, r.ContentLength)
	_, err = r.Body.Read(body)
	if err != nil {
		respond.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	// Декодируем тело запроса в структуру User
	params, err := decoding.DecodingChirp(body)
	if err != nil {
		respond.RespondWithError(w, http.StatusBadRequest, "Invalid user data")
		return
	}

	if len(params.Body) > 140 {
		respond.RespondWithError(w, http.StatusBadRequest, errLong.Error)
		return
	}
	cleanedChirp := domain.Chirp{
		Body:   helpers.CleanBody(params.Body),
		UserID: userId,
	}
	context := r.Context()
	chirp, err := h.service.CreateRowChirp(context, cleanedChirp)
	if err != nil {
		respond.RespondWithError(w, http.StatusInternalServerError, "Error creating chirp")
	}
	respond.RespondWithJSON(w, http.StatusCreated, chirp)
}

func (h Handler) GetOneChirpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	chirpId := r.PathValue("chirpID")
	context := r.Context()
	chirpStruct, err := h.service.GetOneChirp(chirpId, context)
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

	chirps, err := h.service.GetChirpsAuthor(context, authorStr, sortStr)
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
	</html>`, h.cfg.FILESERVERHITS.Load()))
	w.Write(bb)
}

func MiddlewareMetricsInc(cfg *config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FILESERVERHITS.Add(1)

		next.ServeHTTP(w, r)
	})
}

func GetUserId(Cfg *config.Config, header http.Header) (userId uuid.UUID, err error) {
	token, err := auth.GetBearerToken(header)
	if err != nil {
		return userId, err
	}
	userId, err = auth.ValidateJWT(token, Cfg.TOKENSECRET)
	if err != nil {
		return userId, err
	}
	return userId, nil
}
