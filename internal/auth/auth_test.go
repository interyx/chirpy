package auth

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const tokenSecret = "XKhYOoaAFpN4ofJgIw1QSOk5sTUt/JtlqLgTJ4xQ5+X1jg+2Ojzj6dnpI0+Ukz5l QxvP1IV811fZoSkiq8er6Q=="

func TestMakeJWT(t *testing.T) {
	t.Run("Valid JWT Creation", func(t *testing.T) {
		userID := uuid.New()
		expiresIn := time.Hour

		tokenString, err := MakeJWT(userID, tokenSecret, expiresIn)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
			}

			return []byte(tokenSecret), nil
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
			if claims.Issuer != "chirpy" {
				t.Errorf("Issuer claim doesn't match: got %v, want chirpy", claims.Issuer)
			}

		} else {
			t.Error("could not parse token claims")
		}
	})
}

func TestValidateJWT(t *testing.T) {
	t.Run("Valid JWT Validation", func(t *testing.T) {
		userID := uuid.New()
		expiresIn := time.Hour

		tokenString, err := MakeJWT(userID, tokenSecret, expiresIn)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		decoded, err := ValidateJWT(tokenString, tokenSecret)
		if userID != decoded {
			t.Fatalf("Decoded UUID does not match encoded UUID")
		}

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	})

	t.Run("Expired JWT validation", func(t *testing.T) {
		userID := uuid.New()
		old := time.Now().Add(-time.Hour)
		expiresIn := time.Until(old)

		tokenString, err := MakeJWT(userID, tokenSecret, expiresIn)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		_, err = ValidateJWT(tokenString, tokenSecret)

		if err == nil {
			t.Fatalf("An expired token was marked as valid.")
		}
	})

	t.Run("Test GetBearerToken with valid token", func(t *testing.T) {
		headers := make(http.Header)
		headers.Set("Authorization", "Bearer dummy_token_string")

		token, err := GetBearerToken(headers)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if token != "dummy_token_string" {
			t.Errorf("Expected token to be 'dummy_token_string', got %s", token)
		}
	})
}
