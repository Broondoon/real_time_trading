package main

import (
	networkHttp "Shared/network/http"
	networkQueue "Shared/network/queue"
	"auth-service/handlers"
	databaseAccessAuth "databaseAccessAuth"
	"databaseAccessUserManagement"
	"log"
	"os"
)

func main() {
	// Initialize the shared network manager.
	networkManager := networkHttp.NewNetworkHttp()
	networkManagerQueue := networkQueue.NewNetworkQueue(nil, os.Getenv("AUTH_HOST")+":"+os.Getenv("AUTH_PORT"))

	// Create the auth-database access dependency.
	databaseAccess := databaseAccessAuth.NewDatabaseAccess(&databaseAccessAuth.NewDatabaseAccessParams{
		Network: networkManagerQueue,
	})
	databaseAccessWallet := databaseAccessUserManagement.NewDatabaseAccess(&databaseAccessUserManagement.NewDatabaseAccessParams{
		Network: networkManagerQueue,
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
