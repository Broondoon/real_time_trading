package main

import (
	"auth-service/database"
	"auth-service/handlers" // Import handlers package
	"auth-service/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	database.ConnectDatabase()
	r := gin.Default()

	// Public Endpoints
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)

	// Protected Routes
	protected := r.Group("/protected")
	protected.Use(middleware.JWTMiddleware())

	// Add a protected /profile route
	protected.GET("/profile", handlers.Profile)

	r.Run(":8000")
}
