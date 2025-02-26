package handlers

import (
	"auth-service/database"
	"auth-service/models"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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
	err := database.DB.Where("username = ?", input.Username).First(&existingUser).Error
	if err == nil {
		RespondError(c, http.StatusBadRequest, "Username already exists.")
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Only log unexpected errors.
		log.Printf("Error checking for existing username: %v", err)
		RespondError(c, http.StatusInternalServerError, "Internal error")
		return
	}

	// Hash the password.
	hashedPassword, err := HashPassword(input.Password)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "Error hashing password.")
		return
	}

	// Prepare the user model.
	user := models.User{
		Username: input.Username,
		Password: hashedPassword,
		Name:     input.Name,
	}

	// Begin a transaction.
	tx := database.DB.Begin()
	if tx.Error != nil {
		log.Printf("Error starting transaction: %v", tx.Error)
		RespondError(c, http.StatusInternalServerError, "Internal error")
		return
	}

	// Insert the user within the transaction.
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		RespondError(c, http.StatusInternalServerError, "Failed to add user to database.")
		return
	}

	// Construct the URL for the createWallet endpoint.
	umHost := os.Getenv("USER_MANAGEMENT_HOST")
	umPort := os.Getenv("USER_MANAGEMENT_PORT")
	if umHost == "" || umPort == "" {
		tx.Rollback()
		RespondError(c, http.StatusInternalServerError, "User management service not found.")
		return
	}
	walletURL := fmt.Sprintf("http://%s:%s/transaction/createWallet?userID=%s", umHost, umPort, user.ID)

	// Call the createWallet endpoint.
	resp, err := http.Get(walletURL)
	if err != nil {
		tx.Rollback()
		RespondError(c, http.StatusInternalServerError, "Error with wallet creation request.")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		tx.Rollback()
		RespondError(c, resp.StatusCode, string(bodyBytes))
		return
	}

	// Commit the transaction since all steps succeeded.
	if err := tx.Commit().Error; err != nil {
		RespondError(c, http.StatusInternalServerError, "Failed to commit transaction.")
		return
	}

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
