package main

import (
	networkHttp "Shared/network/http"
	"auth-service/handlers"
	databaseAccessAuth "databaseAccessAuth"
	"databaseAccessUserManagement"
	"log"
	"os"
)

func main() {
	// Initialize the shared network manager.
	networkManager := networkHttp.NewNetworkHttp()

	// Create the auth-database access dependency.
	databaseAccess := databaseAccessAuth.NewDatabaseAccess(&databaseAccessAuth.NewDatabaseAccessParams{
		Network: networkManager,
	})
	databaseAccessWallet := databaseAccessUserManagement.NewDatabaseAccess(&databaseAccessUserManagement.NewDatabaseAccessParams{
		Network: networkManager,
	}).Wallet()

	// Inject it into the HTTP handlers.
	go handlers.InitializeUser(databaseAccess, networkManager, databaseAccessWallet)

	//	router := gin.Default()
	//	router.POST("/authentication/register", handlers.Register)
	//	router.POST("/authentication/login", handlers.Login)
	//	router.GET("/authentication/test", handlers.Test)

	log.Printf("Auth-service listening on port %s", os.Getenv("AUTH_PORT"))
	//	http.ListenAndServe(":"+port, router)
	networkManager.Listen()
	<-make(chan struct{})
}
