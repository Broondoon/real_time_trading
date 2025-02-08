package main

import (
	"auth-service/database"
	"auth-service/handlers"
	"auth-service/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	database.ConnectDatabase()

	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)

	protected := r.Group("/protected")
	protected.Use(middleware.JWTAuthMiddleware())
	protected.GET(
		"/profile",
		func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Protected endpoint(s)"})
		})

	r.Run(":8000")
}
