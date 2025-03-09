package main

import (
	networkHttp "Shared/network/http"
	databaseServiceAuth "databaseServiceAuth/database-connection"
	authDatabaseHandlers "databaseServiceAuth/handlers"
	"log"
)

func main() {
	// Establish the database connection.
	// databaseServiceAuth.ConnectDatabase()
	_networkManager := networkHttp.NewNetworkHttp()
	_databaseManager := databaseServiceAuth.NewDatabaseService(&databaseServiceAuth.NewDatabaseServiceParams{})

	// Register the /users endpoint.
	// 	http.HandleFunc("/user", authDatabaseHandlers.GetUserHandler)
	// 	http.HandleFunc("/user/create", authDatabaseHandlers.CreateUserHandler)
	// 	http.HandleFunc("/health", HealthHandler)

	//userAccess := databaseAccess.User()

	go authDatabaseHandlers.InitializeHandlers(_networkManager, _databaseManager)

	log.Println("Auth Database Service Started")

	_networkManager.Listen()
}
