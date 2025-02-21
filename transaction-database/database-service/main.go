package main

import (
	"Shared/network"
	databaseServiceTransaction "databaseServiceTransaction/database-connection"
	transactionDatabaseHandlers "databaseServiceTransaction/handlers"
)

//"Shared/network"

func main() {
	networkManager := network.NewNetwork()
	_databaseManager := databaseServiceTransaction.NewDatabaseService(&databaseServiceTransaction.NewDatabaseServiceParams{})

	go transactionDatabaseHandlers.InitalizeHandlers(networkManager, _databaseManager)
	println("Transaction Database Service Started")

	networkManager.Listen(network.ListenerParams{
		Handler: nil,
	})
}
