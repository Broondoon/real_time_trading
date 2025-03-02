package main

import (
	"Shared/network"
	databaseAccessAuth "auth-database/database-access"
	"auth-service/handlers" // Import handlers package
	"auth-service/middleware"

	"github.com/gin-gonic/gin"
)

func main() {

	networkManager := network.NewNetwork()
	databaseAccess := databaseAccessAuth.NewDatabaseAccess(&databaseAccessAuth.NewDatabaseAccessParams{
		Network: networkManager,
	})

	databaseAccess.Connect()
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
