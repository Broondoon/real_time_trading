package main

import (
	"Shared/network"
	"auth-service/handlers"
	databaseAccessAuth "databaseAccessAuth"
	"log"
	"os"
)

func main() {
	// Initialize the shared network manager.
	networkManager := network.NewNetwork()

	// Create the auth-database access dependency.
	databaseAccess := databaseAccessAuth.NewDatabaseAccess(&databaseAccessAuth.NewDatabaseAccessParams{
		Network: networkManager,
	})

	userAccess := databaseAccess.User()
	// Inject it into the HTTP handlers.
	handlers.InitializeUser(userAccess, networkManager)

	//	router := gin.Default()
	//	router.POST("/authentication/register", handlers.Register)
	//	router.POST("/authentication/login", handlers.Login)
	//	router.GET("/authentication/test", handlers.Test)

	port := os.Getenv("AUTH_SERVICE_PORT")
	if port == "" {
		port = "8000"
	}
	log.Printf("Auth-service listening on port %s", port)
	//	http.ListenAndServe(":"+port, router)
	networkManager.Listen(network.ListenerParams{Handler: nil})
}
