package tokens

import (
	"Chirpy/internal/auth"
	"time"

	"github.com/google/uuid"
)

type Saved struct {
	Token        string
	RefreshToken string
	ExpiresAt    time.Time
}

func CheckPassword(password string, hashedPassword string) error {
	errCheck := auth.CheckPasswordHash(password, hashedPassword)
	if errCheck != nil {
		return errCheck
	}
	return nil
}

func GetToken(id uuid.UUID, tokenSecret string) (tk Saved, err error) {
	tk.Token, err = auth.MakeJWT(id, tokenSecret)
	if err != nil {
		return Saved{}, err
	}
	tk.RefreshToken, err = auth.MakeRefreshToken()
	if err != nil {
		return Saved{}, err
	}
	tk.ExpiresAt = time.Now().Add(60 * 24 * time.Hour)

	return tk, nil
}
