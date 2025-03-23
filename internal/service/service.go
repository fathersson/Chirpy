package service

import (
	"Chirpy/config"
	"Chirpy/domain"
	"Chirpy/internal/auth"
	"Chirpy/internal/database"
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

// Интерфейс для базы данных
type Database interface {
	CreateUser(context.Context, database.CreateUserParams) (database.User, error)
	CreateChirp(context.Context, database.CreateChirpParams) (database.Chirp, error)
	GetChirpsAuthorID(context.Context, uuid.UUID) ([]database.Chirp, error)
	GetChirp(context.Context, uuid.UUID) (database.Chirp, error)
	CheckUser(ctx context.Context, id uuid.UUID) (database.User, error)
	UpdateChirpyRed(ctx context.Context, id uuid.UUID) (database.User, error)
	DeleteChirp(ctx context.Context, id uuid.UUID) error
	DeleteUsers(ctx context.Context) error
	UpdateEMailPasword(ctx context.Context, arg database.UpdateEMailPaswordParams) (database.User, error)
	RevokeToken(ctx context.Context, token string) error
	GetUserFromRefreshToken(ctx context.Context, token string) (database.RefreshToken, error)
	GetUser(ctx context.Context, email string) (database.User, error)
	CreateRefreshToken(ctx context.Context, arg database.CreateRefreshTokenParams) (database.RefreshToken, error)
	//QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
}

type DbUser interface {
	GetUserId(Cfg *config.Config, header http.Header) (userId uuid.UUID, err error)
	CreateRowUser(context context.Context, returnParams database.User) (user domain.User, err error)
	CreateRowChirp(context context.Context, cleanedChirp database.Chirp) (chirp domain.Chirp, err error)
	GetChirpsAuthor(context context.Context, authorStr string, sortOrder string) (chirpStruct []domain.Chirp, err error)
	GetOneChirp(chirpId string, context context.Context) (chirp domain.Chirp, err error)
}

type ServiceUser struct {
	DbU DbUser
}

func NewServiceUser() *ServiceUser {
	return &ServiceUser{}
}

// Структура сервиса, которая использует интерфейс Database
type Service struct {
	Db Database
}

// Конструктор сервиса, который принимает интерфейс Database
func NewService(db Database) *Service {
	return &Service{Db: db}
}

func (s *Service) GetUserId(Cfg *config.Config, header http.Header) (userId uuid.UUID, err error) {
	token, err := auth.GetBearerToken(header)
	if err != nil {
		return userId, err
	}
	userId, err = auth.ValidateJWT(token, Cfg.TokenSecret)
	if err != nil {
		return userId, err
	}
	return userId, nil
}

func (s *Service) CreateRowUser(context context.Context, returnParams database.User) (user domain.User, err error) {
	CreateHash(&returnParams)
	createUserParams := database.CreateUserParams{
		Email:          returnParams.Email,
		HashedPassword: returnParams.HashedPassword,
	}
	userD, err := s.Db.CreateUser(context, createUserParams)
	if err != nil {
		return domain.User{}, err
	}
	return ToUser(userD), nil
}

func (s *Service) CreateRowChirp(context context.Context, cleanedChirp database.Chirp) (chirp domain.Chirp, err error) {
	ChirpParams := database.CreateChirpParams{
		Body:   cleanedChirp.Body,
		UserID: cleanedChirp.UserID,
	}
	chirpD, err := s.Db.CreateChirp(context, ChirpParams)
	if err != nil {
		return domain.Chirp{}, err
	}
	return ToChirp(chirpD), nil
}

/*func (s *Service) GetChirpsAuthor(context context.Context, authorStr string, sortOrder string) (chirpStruct []domain.Chirp, err error) {
	userId, err := uuid.Parse(authorStr)
	if err != nil {
		return []domain.Chirp{}, err
	}

	//chirps, err := s.db.GetChirpsAuthorID(context, userId)
	//if err != nil {
	//	return []domain.Chirp{}, err
	//}
	//chirpStruct = chirpsForStruct(chirps)
	//return chirpStruct, nil
	// В зависимости от параметра sortOrder формируем порядок сортировки

	var chirps []database.Chirp
	var orderClause string
	if sortOrder == "desc" {
		orderClause = "ORDER BY created_at DESC"
	} else {
		orderClause = "ORDER BY created_at ASC"
	}

	// Выполняем запрос с сортировкой
	query := fmt.Sprintf("SELECT * FROM chirps WHERE user_id = $1 %s", orderClause)
	rows, err := s.Db.QueryContext(context, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Обработка результата запроса
	for rows.Next() {
		var chirp database.Chirp
		if err := rows.Scan(&chirp.ID, &chirp.Body, &chirp.UserID, &chirp.CreatedAt); err != nil {
			return nil, err
		}
		chirps = append(chirps, chirp)
	}

	return chirpsForStruct(chirps), nil
}*/

func (s *Service) GetOneChirp(chirpId string, context context.Context) (chirp domain.Chirp, err error) {
	chirpParse, err := uuid.Parse(chirpId)
	if err != nil {
		return domain.Chirp{}, err
	}

	dbChirp, err := s.Db.GetChirp(context, chirpParse)
	if err != nil {
		return domain.Chirp{}, err
	}
	chirpStruct := ToChirp(dbChirp)
	return chirpStruct, nil
}

func CreateHash(returnParams *database.User) {
	var err error
	returnParams.HashedPassword, err = auth.HashPassword(returnParams.HashedPassword)
	if err != nil {
		slog.Error(fmt.Sprintf("%v", err))
		return
	}
}

func ToChirp(dbChirp database.Chirp) domain.Chirp {
	return domain.Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}
}

func ToUser(dbUser database.User) domain.User {
	return domain.User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}
}

func ToDbUser(user domain.User) database.User {
	return database.User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
}

func chirpsForStruct(chirps []database.Chirp) (chirpStruct []domain.Chirp) {
	for _, v := range chirps {
		chirpStruct = append(chirpStruct, ToChirp(v))
	}
	return
}
