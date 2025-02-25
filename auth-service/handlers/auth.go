package handlers

import (
	"auth-service/database"
	"auth-service/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"time"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateToken(userID uint) (string, error) {
	var jwtsecret = []byte(os.Getenv("JWT_SECRET"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 1).Unix(),
	})

	return token.SignedString(jwtsecret)
}

func Register(c *gin.Context) {
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
		Username: input.Username,
		Password: hashedPassword,
		Name:     input.Name,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		RespondError(c, http.StatusInternalServerError, err.Error())
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

	if user.ID == 0 || !CheckPasswordHash(input.Password, user.Password) {
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
