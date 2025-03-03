package networkHttp

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func generateTestToken() string {
	// Set a test secret
	os.Setenv("JWT_SECRET", "supersecretkey")
	jwtSecret := os.Getenv("JWT_SECRET")

	// Create token with user ID = 123 for testing
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": 123,
		"exp": time.Now().Add(time.Hour * 1).Unix(),
	})

	tokenString, _ := token.SignedString([]byte(jwtSecret))
	return tokenString
}

func TestExtractUserIDFromToken(t *testing.T) {
	token := generateTestToken() // Generate a valid token

	userID, err := ExtractUserIDFromToken(token)
	if err != nil {
		t.Fatalf("Failed to extract user ID: %v", err)
	}

	// if userID != 123 {
	// 	t.Fatalf("Expected user ID 123, but got %d", userID)
	// }

	t.Logf("Extracted user ID successfully: %d", userID)
}
