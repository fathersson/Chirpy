package auth

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestJWTCreationAndValidation(t *testing.T) {
	// Let's start with a test user ID and secret
	userID := uuid.New()
	secret := "test_secret"

	// What do you think we should test first?
	// Hint: We should create a token and then validate it
	t.Run("valid token", func(t *testing.T) {
		token, err := MakeJWT(userID, secret)
		if err != nil {
			t.Fatalf("Error creating token: %v", err)
		}

		gotUID, err := ValidateJWT(token, secret)
		if err != nil {
			t.Fatalf("Error validating token: %v", err)
		}

		if gotUID != userID {
			t.Errorf("Expected user ID %v, got %v", userID, gotUID)
		}
	})

	/*t.Run("expired token", func(t *testing.T) {
		// How would you test an expired token?
		// Hint: Use a very short duration
		token, err := MakeJWT(userID, secret, time.Millisecond*1)
		if err != nil {
			t.Fatalf("Error creating token: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
		_, err = ValidateJWT(token, secret)
		if err == nil {
			t.Error("Expected error for expired token, got nil")
		}
	})*/

	t.Run("invalid secret", func(t *testing.T) {
		// How would you test a token with wrong secret?
		// Hint: Create with one secret, validate with another
		token, err := MakeJWT(userID, secret)
		if err != nil {
			t.Fatalf("Error creating token: %v", err)
		}

		wrongsecret := "wrong_secret"
		_, err = ValidateJWT(token, wrongsecret)
		if err == nil {
			t.Error("Expected error for invalid secret, got nil") // More descriptive error message
		}
	})
}

func TestGetBearerToken(t *testing.T) {
	//check GetBearerToken
	t.Run("Valid GetBearerToken", func(t *testing.T) {
		header := make(http.Header)
		header.Add("Authorization", "Bearer valid-token")
		token, err := GetBearerToken(header)
		if err != nil {
			t.Fatalf("Error get token: %v", err)
		}
		if token != "valid-token" {
			t.Fatalf("Expected token 'valid-token', got '%s'", token)
		}
	})

	t.Run("Wrong prefix", func(t *testing.T) {
		header := make(http.Header)
		header.Add("Authorization", "wrong prefix valid token")
		_, err := GetBearerToken(header)
		if err == nil {
			t.Fatalf("Expected error for wrong prefix, got nil: %v", err)
		}
	})

	t.Run("Empty token", func(t *testing.T) {
		header := make(http.Header)
		header.Add("Authorization", "Bearer ")
		_, err := GetBearerToken(header)
		if err == nil {
			t.Fatalf("Expected error for empty token, got nil: %v", err)
		}
	})

	t.Run("Empty header", func(t *testing.T) {
		header := make(http.Header)
		header.Add("Authorization", "")
		_, err := GetBearerToken(header)
		if err == nil {
			t.Fatalf("Expected error for empty header, got nil: %v", err)
		}
	})

	t.Run("Absent header", func(t *testing.T) {
		header := make(http.Header)
		_, err := GetBearerToken(header)
		if err == nil {
			t.Fatalf("Expected error for absent header, got nil: %v", err)
		}
	})
}
