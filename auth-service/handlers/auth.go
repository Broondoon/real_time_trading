package handlers

import (
	"auth-service/database"
	"auth-service/models"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

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

func Register(c *gin.Context) {
	// Bind incoming JSON to our User model.
	var input models.User
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	// Check if the username already exists.
	var existingUser models.User
	if err := database.DB.Where("Username = ?", input.Username).First(&existingUser).Error; err == nil {
		RespondError(c, http.StatusBadRequest, "User already exists.")
		return
	}

	// Hash the password.
	hashedPassword, err := HashPassword(input.Password)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "Error hasning password.")
		return
	}
	user := models.User{
		ID:       uuid.New().String(),
		Username: input.Username,
		Password: hashedPassword,
		Name:     input.Name,
	}

	// Add the user to the database.
	if err := database.DB.Create(&user).Error; err != nil {
		RespondError(c, http.StatusBadRequest, "Registration failed")
		return
	}

	// Construct the URL for the createWallet endpoint.
	umHost := os.Getenv("USER_MANAGEMENT_HOST")
	umPort := os.Getenv("USER_MANAGEMENT_PORT")
	if umHost == "" || umPort == "" {
		RespondError(c, http.StatusInternalServerError, "User management service not configured")
		return
	}
	fmt.Println("id:", user.ID)
	//// Build the URL with a query parameter for the user ID.
	walletURL := fmt.Sprintf("http://%s:%s/createWallet?userID=%s", umHost, umPort, user.ID)
	// Send a GET request to create the wallet.
	resp, err := http.Get(walletURL)
	fmt.Println("Resp:", resp)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "Failed to create wallet")
		return
	}
	defer resp.Body.Close()

	// If the wallet creation didn't return a 200 OK, return its error message
	fmt.Println("Resp:", resp)
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		RespondError(c, http.StatusBadRequest, string(bodyBytes))
		//    RespondError(c, resp.StatusCode, string(bodyBytes))
		return
	}
	fmt.Println("Resp:", resp)
	// Everything succeeded; return a success response.
	RespondSuccess(c, nil)
}

func Login(c *gin.Context) {
	var input models.User
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	var user models.User
	database.DB.Where("username = ?", input.Username).First(&user)
	if user.ID == "" || !CheckPasswordHash(input.Password, user.Password) {

		RespondError(c, http.StatusBadRequest, "Invalid Credentials.")
		return
	}

	token, err := GenerateToken(user.ID)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "Token generation failed.")
		return
	}

	RespondSuccess(c, gin.H{
		"token": token,
	})
}

func RespondSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

func RespondError(c *gin.Context, statusCode int, errorMsg string) {
	c.JSON(statusCode, gin.H{
		"success": false,
		"error":   errorMsg,
	})
}

func Test(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "You have successfully queried a " +
		"protected endpoint with your JWT token. Excellent!",
		"userID": userID,
	})
}
