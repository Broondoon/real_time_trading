package networkHttp

import (
	"Shared/network"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// HttpClientInterface is an interface for the HttpClient struct

type HttpClient struct {
	BaseURL   string
	AuthToken string
	Client    *http.Client
	SecretKey []byte
}

func (hc *HttpClient) GetBaseURL() string {
	return hc.BaseURL
}

func newHttpClient(baseURL string) network.ClientInterface {
	return &HttpClient{
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (hc *HttpClient) authenticate(req *http.Request) error {
	if hc.AuthToken == "" {
		return errors.New("no token found, authentication required")
	}
	req.Header.Set("token", fmt.Sprintf("Bearer %s", hc.AuthToken))
	return nil
}

func (hc *HttpClient) handleResponse(resp *http.Response) ([]byte, error) {
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server returned error: %d %s", resp.StatusCode, resp.Status)
	}
	if resp.StatusCode == http.StatusResetContent {
		return nil, errors.New("204 No Content")
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (hc *HttpClient) Get(endpoint string, queryParams map[string]string) ([]byte, error) {
	url, err := url.Parse(hc.BaseURL + endpoint)
	fmt.Printf("[DEBUG] GET Request URL: %s\n", url.String())
	if err != nil {
		return nil, err
	}

	q := url.Query()
	for key, value := range queryParams {
		q.Add(key, value)
	}
	url.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	// if err := hc.authenticate(req); err != nil {
	// 	return nil, err
	// }

	resp, err := hc.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return hc.handleResponse(resp)
}

func (hc *HttpClient) PostBulk(endpoint string, payload []interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("DEBUG: Error marshalling payload:", err.Error())
		return nil, err
	}
	fmt.Printf("DEBUG: Payload marshalled successfully: %s\n", string(jsonData))

	fullURL := hc.BaseURL + endpoint
	req, err := http.NewRequest(http.MethodPost, fullURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("DEBUG: Error creating POST request:", err.Error())
		return nil, err
	}
	fmt.Printf("DEBUG: Created POST request for URL: %s\n", fullURL)

	req.Header.Set("Content-Type", "application/json")
	// if err := hc.authenticate(req); err != nil {
	// 	return nil, err
	// }
	req.Header.Set("isBulk", "true")

	fmt.Println("DEBUG: Sending POST request...")
	resp, err := hc.Client.Do(req)
	if err != nil {
		fmt.Println("DEBUG: Error sending POST request:", err.Error())
		return nil, err
	}
	fmt.Printf("DEBUG: Received response with status: %s\n", resp.Status)

	return hc.handleResponse(resp)

}

func (hc *HttpClient) Post(endpoint string, payload interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("DEBUG: Error marshalling payload:", err.Error())
		return nil, err
	}
	fmt.Printf("DEBUG: Payload marshalled successfully: %s\n", string(jsonData))

	fullURL := hc.BaseURL + endpoint
	req, err := http.NewRequest(http.MethodPost, fullURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("DEBUG: Error creating POST request:", err.Error())
		return nil, err
	}
	fmt.Printf("DEBUG: Created POST request for URL: %s\n", fullURL)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("isBulk", "false")
	// if err := hc.authenticate(req); err != nil {
	// 	return nil, err
	// }

	fmt.Println("DEBUG: Sending POST request...")
	resp, err := hc.Client.Do(req)
	if err != nil {
		fmt.Println("DEBUG: Error sending POST request:", err.Error())
		return nil, err
	}
	fmt.Printf("DEBUG: Received response with status: %s\n", resp.Status)

	return hc.handleResponse(resp)
}

func (hc *HttpClient) Put(endpoint string, payload interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, hc.BaseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	// if err := hc.authenticate(req); err != nil {
	// 	return nil, err
	// }

	resp, err := hc.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return hc.handleResponse(resp)
}

func (hc *HttpClient) Delete(endpoint string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodDelete, hc.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	// if err := hc.authenticate(req); err != nil {
	// 	return nil, err
	// }

	resp, err := hc.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return hc.handleResponse(resp)
}

// ExtractUserIDFromToken extracts the user ID from a JWT token
func ExtractUserIDFromToken(tokenString string) (string, error) {
	if tokenString == "" {
		log.Println("[ExtractUserIDFromToken] Missing token in request")
		return "", errors.New("missing token in request")
	}

	// Remove "Bearer " prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// Ensure JWT_SECRET is loaded
	if len(jwtSecret) == 0 {
		log.Println("[ExtractUserIDFromToken] JWT secret is missing")
		return "", errors.New("server misconfiguration: JWT secret is missing")
	}

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("[ExtractUserIDFromToken] Unexpected signing method: %v", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		log.Printf("[ExtractUserIDFromToken] Token parsing error: %v", err)
		return "", fmt.Errorf("invalid token: %v", err)
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("[ExtractUserIDFromToken] Failed to parse token claims")
		return "", errors.New("invalid claims structure in token")
	}

	// Extract userID from claims
	userID, ok := claims["sub"].(string)
	if !ok {
		log.Println("[ExtractUserIDFromToken] Missing or malformed user ID in token")
		return "", errors.New("missing or malformed user ID in token")
	}

	return userID, nil
}

// contextKey is a type for context keys to avoid key collisions.
type contextKey string

// userIDKey is the key used for storing user ID in the context.
var userIDKey = contextKey("userID")

func TokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("token")
		if tokenString == "" {
			http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
			return
		}
		//Validate token and extract user ID
		userID, err := ExtractUserIDFromToken(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}
		//Optionally, you can add the userID to the context:
		//userID := "6fd2fc6b-9142-4777-8b30-575ff6fa2460"
		ctx := r.Context()
		ctx = context.WithValue(ctx, userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
