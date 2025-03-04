package handlers

import (
	user "Shared/entities/user"
	"Shared/network"
	databaseAccessAuth "databaseAccessAuth"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ---------- Utility Functions ----------

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateToken(userID string) (string, error) {
	var jwtsecret = []byte(os.Getenv("JWT_SECRET"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 1).Unix(),
	})
	return token.SignedString(jwtsecret)
}

// ---------- Dependency Injection ----------

// authDB is the dependency injected from main.go.
// It implements databaseAccessAuth.AuthDataAccessInterface.
var _authDB databaseAccessAuth.UserDataAccessInterface

// InitializeAuthHandlers sets up the dependency for the handlers.
func InitializeUser(db databaseAccessAuth.UserDataAccessInterface, networkManager network.NetworkInterface) {
	_authDB = db
	networkManager.AddHandleFuncProtected(network.HandlerParams{Pattern: "authentication/register", Handler: Register})
	networkManager.AddHandleFuncProtected(network.HandlerParams{Pattern: "authentication/login", Handler: Login})
}

// ---------- HTTP Handlers ----------

// ---------- HTTP Handlers ----------

// Register handles user registration.
// It expects a POST request with a JSON body representing a user.
func Register(responseWriter http.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	// Decode the JSON body into a user.
	var input user.User
	if err := json.Unmarshal(data, &input); err != nil {
		http.Error(responseWriter, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Check if the username already exists.
	existingUser, err := _authDB.GetUserByUsername(input.GetUsername())
	if err == nil && existingUser != nil {
		http.Error(responseWriter, "Username already exists.", http.StatusBadRequest)
		return
	} else if err != nil && err.Error() != "user not found" {
		// Log unexpected errors.
		http.Error(responseWriter, "Internal error", http.StatusInternalServerError)
		return
	}

	// (Optional debug printing if needed)
	// If existingUser is nil (which is expected when user is not found), skip JSON conversion.
	// Otherwise, you might log the user details.

	// Hash the password.
	hashedPassword, err := HashPassword(input.GetPassword())
	if err != nil {
		http.Error(responseWriter, "Error hashing password.", http.StatusInternalServerError)
		return
	}
	input.Password = hashedPassword

	// Create the user using the new interface.
	if err := _authDB.CreateUser(&input); err != nil {
		http.Error(responseWriter, "Failed to add user to database.", http.StatusInternalServerError)
		return
	}

	// Optionally, call a wallet creation endpoint.
	umHost := os.Getenv("USER_MANAGEMENT_HOST")
	umPort := os.Getenv("USER_MANAGEMENT_PORT")
	if umHost == "" || umPort == "" {
		http.Error(responseWriter, "User management service not found.", http.StatusInternalServerError)
		return
	}
	walletURL := fmt.Sprintf("http://%s:%s/transaction/createWallet?userID=%s", umHost, umPort, input.GetId())
	resp, err := http.Get(walletURL)
	if err != nil {
		http.Error(responseWriter, "Error with wallet creation request.", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		http.Error(responseWriter, string(bodyBytes), resp.StatusCode)
		return
	}

	// Return a success response.
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusCreated)
	json.NewEncoder(responseWriter).Encode(map[string]interface{}{
		"success": true,
		"data":    nil,
	})
}

// Login handles user login.
// It expects a POST request with a JSON body containing username and password.
func Login(responseWriter http.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	// Decode the JSON body.
	var input user.User
	if err := json.Unmarshal(data, &input); err != nil {
		http.Error(responseWriter, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Retrieve the user by username.
	u, err := _authDB.GetUserByUsername(input.GetUsername())
	if err != nil || u == nil || !CheckPasswordHash(input.GetPassword(), u.GetPassword()) {
		http.Error(responseWriter, "Invalid Credentials.", http.StatusBadRequest)
		return
	}

	// Generate a JWT token.
	token, err := GenerateToken(u.GetId())
	if err != nil {
		http.Error(responseWriter, "Token generation failed.", http.StatusInternalServerError)
		return
	}

	// Return the token in a success JSON response.
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusOK)
	json.NewEncoder(responseWriter).Encode(map[string]interface{}{
		"success": true,
		"token":   token,
	})
}
