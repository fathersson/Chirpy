package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashPasw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return string(hashPasw), nil
}

func CheckPasswordHash(password, hashPasw string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashPasw), []byte(password))
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string) (string, error) {
	mySigningKey := []byte(tokenSecret)
	expiration := time.Now().Add(time.Hour).Unix()

	claims := jwt.MapClaims{
		"sub": userID,     // subject (ID пользователя)
		"exp": expiration, // expiration time
		"iat": jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(mySigningKey)
	return ss, err
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	mySigningKey := []byte(tokenSecret)

	// Парсим токен с claims
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Возвращаем ключ для подписи
		return mySigningKey, nil
	})

	if err != nil {
		return uuid.UUID{}, err
	}

	// Проверяем, действителен ли токен
	if _, ok := token.Claims.(*jwt.RegisteredClaims); !ok || !token.Valid {
		return uuid.UUID{}, fmt.Errorf("invalid token")
	}
	ui, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid user ID in token: %v", err)
	}
	return ui, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("Authorization header is missing")
	}
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("Authorization header is missing or invalid")
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	return token, nil
}

func MakeRefreshToken() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	encodedKey := hex.EncodeToString(key)
	return encodedKey, nil
}

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("Authorization header is missing")
	}
	if !strings.HasPrefix(authHeader, "ApiKey ") {
		return "", fmt.Errorf("Invalid Authorization header format")
	}
	key := strings.TrimPrefix(authHeader, "ApiKey ")
	return key, nil
}
