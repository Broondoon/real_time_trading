package main

import (
	networkHttp "Shared/network/http"
	"log"
	mongoDatabaseServiceAuth "mongoDatabaseServiceAuth/database-connection"
	authMongoDatabaseHandlers "mongoDatabaseServiceAuth/handlers"
)

func main() {
	// Establish the database connection.
	_networkManager := networkHttp.NewNetworkHttp()
	_databaseManager := mongoDatabaseServiceAuth.NewDatabaseService(&mongoDatabaseServiceAuth.NewDatabaseServiceParams{})

	go authMongoDatabaseHandlers.InitializeHandlers(_networkManager, _databaseManager)

	log.Println("Auth Database Service Started")

	_networkManager.Listen()
}
