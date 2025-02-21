package main

import (
	OrderInitiatorService "OrderInitiatorService/handlers"
	"Shared/network"
	"databaseAccessTransaction"
)

//"Shared/network"

func main() {
	//Need to upgrade to use my entity class stuff and the new services.

	networkManager := network.NewNetwork()
	databaseAccess := databaseAccessTransaction.NewDatabaseAccess(&databaseAccessTransaction.NewDatabaseAccessParams{})

	go OrderInitiatorService.InitalizeHandlers(networkManager, databaseAccess)
	println("Matching Engine Service Started")

	networkManager.Listen(network.ListenerParams{
		Handler: nil,
	})
}
