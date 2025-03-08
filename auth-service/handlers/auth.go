package handlers

import (
	user "Shared/entities/user"
	"Shared/network"
	databaseAccessAuth "databaseAccessAuth"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

// ---------- Response Helpers ----------

func RespondSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"success": true,
		"data":    data,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func RespondError(w http.ResponseWriter, statusCode int, errorMsg string) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"success": false,
		"error":   errorMsg,
	}
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// ---------- Dependency Injection ----------

// authDB is the dependency injected from main.go.
// It implements databaseAccessAdduth.AuthDataAccessInterface.
var _authDB databaseAccessAuth.UserDataAccessInterface

// InitializeAuthHandlers sets up the dependency for the handlers.
func InitializeUser(db databaseAccessAuth.UserDataAccessInterface, networkManager network.NetworkInterface) {
	_authDB = db
	networkManager.AddHandleFuncUnprotected(network.HandlerParams{Pattern: "authentication/register", Handler: Register})
	networkManager.AddHandleFuncUnprotected(network.HandlerParams{Pattern: "authentication/login", Handler: Login})
}

// ---------- HTTP Handlers ----------
// Register handles user registration.
func Register(w network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	log.Println("Register() called by handler in Auth-service.")

	// Decode the JSON body into a User object.
	var input user.User
	if err := json.Unmarshal(data, &input); err != nil {
		RespondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid JSON: %v", err))
		return
	}

	// Check if the username already exists.
	existingUser, err := _authDB.GetUserByUsername(input.GetUsername())
	if existingUser != nil {
		log.Printf("Username exists?: %s", existingUser)
		RespondError(w, http.StatusBadRequest, "Username already exists.")
		return
	} else if err != nil && err.Error() != "user not found" {
		// Unexpected error.
		log.Printf("Couldn't find user? %s", err.Error())
		RespondError(w, http.StatusInternalServerError, "Internal error")
		return
	}

	// Hash the password.
	hashedPassword, err := HashPassword(input.GetPassword())
	if err != nil {
		log.Printf("error hashing: %s", err)
		RespondError(w, http.StatusInternalServerError, "Error hashing password.")
		return
	}
	input.Password = hashedPassword

	// Create the user.
	if err := _authDB.CreateUser(&input); err != nil {
		log.Printf("Failed to add user to database: %s", err)
		RespondError(w, http.StatusInternalServerError, "Failed to add user to database.")
		return
	}
	//TODO: Checking if the user exists, then creating, and then fetching
	// the user from the DB again is not very efficient.
	getUser, err := _authDB.GetUserByUsername(input.Username)

	// Call the wallet creation endpoint.
	umHost := os.Getenv("USER_MANAGEMENT_HOST")
	umPort := os.Getenv("USER_MANAGEMENT_PORT")
	if umHost == "" || umPort == "" {
		log.Printf("Host and Port:: %s:%s", umHost, umPort)
		RespondError(w, http.StatusInternalServerError, "User management service not found.")
		return
	}

	walletURL := fmt.Sprintf("http://%s:%s/transaction/createWallet?userID=%s", umHost, umPort, getUser.GetId())
	log.Printf("This is the input obj: %s", getUser.GetId())
	resp, err := http.Get(walletURL)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Error with wallet creation request.")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		RespondError(w, resp.StatusCode, string(bodyBytes))
		return
	}

	// Return a success response.
	RespondSuccess(w, nil)
}

// Login handles user login.
func Login(w network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	var input user.User
	if err := json.Unmarshal(data, &input); err != nil {
		RespondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid JSON: %v", err))
		return
	}

	u, err := _authDB.GetUserByUsername(input.GetUsername())
	if err != nil || u == nil || !CheckPasswordHash(input.GetPassword(), u.GetPassword()) {
		RespondError(w, http.StatusBadRequest, "Invalid Credentials.")
		return
	}

	token, err := GenerateToken(u.GetId())
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Token generation failed.")
		return
	}

	RespondSuccess(w, map[string]interface{}{"token": token})
}
