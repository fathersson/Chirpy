package domain

import (
	"time"

	"github.com/google/uuid"
)

type Env struct {
	DbUrl       string `env:"DB_URL"`
	Platform    string `env:"PLATFORM"`
	TokenSecret string `env:"TOKEN_SECRET"`
	PolkaKey    string `env:"POLKA_KEY"`
}

type jsonStruct struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

type ErrStruct struct {
	Error string `json:"error"`
}

type valid struct {
	Valid bool `json:"valid"`
}

type cleanStruct struct {
	Clean string `json:"cleaned_body"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Expires   int       `json:"expires"`
	ChirpyRed bool      `json:"is_chirpy_red"`
}

type Webhook struct {
	Event string `json:"event"`
	Data  struct {
		UserID uuid.UUID `json:"user_id"`
	} `json:"data"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

type LoginResponse struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ChirpyRed    bool      `json:"is_chirpy_red"`
}
