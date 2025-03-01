package main

import (
	database "auth-database/database-service"
	"auth-service/handlers" // Import handlers package
	"auth-service/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	database.ConnectDatabase()
	r := gin.Default()

	// Public Endpoints
	auth := r.Group("/authentication")
	auth.POST("/register", handlers.Register)
	auth.POST("/login", handlers.Login)

	// Protected Routes
	protected := r.Group("/protected")
	protected.Use(middleware.JWTMiddleware())

	// Add a protected /profile route
	protected.GET("/test", handlers.Test)

	r.Run(":8000")
}
