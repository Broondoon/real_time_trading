package main

import (
	"Shared/network"
	databaseServiceAuth "databaseServiceAuth/database-connection"
	authDatabaseHandlers "databaseServiceAuth/handlers"
	"fmt"
)

func main() {
	// Establish the database connection.
	// databaseServiceAuth.ConnectDatabase()
	_networkManager := network.NewNetwork()
	_databaseManager := databaseServiceAuth.NewDatabaseService(&databaseServiceAuth.NewDatabaseServiceParams{})

	// Register the /users endpoint.
	// 	http.HandleFunc("/user", authDatabaseHandlers.GetUserHandler)
	// 	http.HandleFunc("/user/create", authDatabaseHandlers.CreateUserHandler)
	// 	http.HandleFunc("/health", HealthHandler)

	//userAccess := databaseAccess.User()

	go authDatabaseHandlers.InitializeHandlers(_networkManager, _databaseManager)

	fmt.Println("Auth Database Service Started")

	_networkManager.Listen(network.ListenerParams{
		Handler: nil,
	})
}
