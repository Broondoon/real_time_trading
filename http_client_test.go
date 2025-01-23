package shared

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockResponse struct {
	Message string `json:"message"`
}

func TestHttpClient_Get(t *testing.T) {

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test_token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		response := MockResponse{Message: "Hello, world!"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	client := NewHttpClient(mockServer.URL)
	client.setAuthToken("test_token")

	queryParams := map[string]string{
		"param1": "value1",
		"param2": "value2",
	}
	responseBody, err := client.Get("/test-endpoint", queryParams)
	if err != nil {
		t.Fatalf("Get request failed: %v", err)
	}

	var response MockResponse
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if response.Message != "Hello, world!" {
		t.Errorf("Expected message 'Hello, world!', got '%s'", response.Message)
	}
}

func TestHttpClient_Get_WithoutToken(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}))
	defer mockServer.Close()

	client := NewHttpClient(mockServer.URL)

	_, err := client.Get("/test-endpoint", nil)
	if err == nil || err.Error() != "no token found, authentication required" {
		t.Fatalf("Expected authentication error, got: %v", err)
	}
}
