package main

import (
	"Shared/network"
	databaseServiceStock "databaseServiceStock/database-connection"
	stockDatabaseHandlers "databaseServiceStock/handlers"
	"fmt"
)

//"Shared/network"

func main() {
	networkManager := network.NewNetwork()
	_databaseManager := databaseServiceStock.NewDatabaseService(&databaseServiceStock.NewDatabaseServiceParams{})

	go stockDatabaseHandlers.InitalizeHandlers(networkManager, _databaseManager)
	fmt.Println("Matching Engine Service Started")

	networkManager.Listen(network.ListenerParams{
		Handler: nil,
	})
}
