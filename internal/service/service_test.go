package service

import (
	"Chirpy/internal/database"
	"Chirpy/internal/domain"
	"Chirpy/internal/service/mocks"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetOneChirp(t *testing.T) {
	mockDB := new(mocks.Database)
	svc := &Db{Db: mockDB}

	validID := uuid.New()
	mockChirp := database.Chirp{
		ID:     validID,
		Body:   "Тестовый чирп",
		UserID: uuid.New(),
	}

	errorID := uuid.New()
	tests := []struct {
		name        string
		chirpID     string
		want        domain.Chirp
		wantErr     bool
		prepareMock func()
	}{
		{
			name:    "успешное получение",
			chirpID: validID.String(),
			want:    ToChirp(mockChirp),
			wantErr: false,
			prepareMock: func() {
				mockDB.On("GetChirp", context.Background(), validID).Return(mockChirp, nil)
			},
		},
		{
			name:        "неверный ID",
			chirpID:     "неверный-id",
			wantErr:     true,
			prepareMock: func() {},
		},
		{
			name:    "ошибка БД",
			chirpID: errorID.String(),
			wantErr: true,
			prepareMock: func() {
				mockDB.On("GetChirp", context.Background(), errorID).Return(database.Chirp{}, errors.New("ошибка БД"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepareMock()

			got, err := svc.GetOneChirp(tt.chirpID, context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
