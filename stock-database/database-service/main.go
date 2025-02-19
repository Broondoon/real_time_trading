package main

import (
	"Shared/database/database-service"
	"Shared/network"
	"databaseServiceStock/database-connection"
	"databaseServiceStock/handlers"
)

//"Shared/network"

func main() {
	networkManager := network.NewNetwork()
	_databaseManager := databaseServiceStock.NewDatabaseService(&databaseServiceStock.NewDatabaseServiceParams{NewPostGresDatabaseParams: &database.NewPostGresDatabaseParams{NewBaseDatabaseParams: &database.NewBaseDatabaseParams{}}})

	go stockDatabaseHandlers.InitalizeHandlers(networkManager, _databaseManager)
	println("Matching Engine Service Started")

	networkManager.Listen(network.ListenerParams{
		Handler: nil,
	})
}
