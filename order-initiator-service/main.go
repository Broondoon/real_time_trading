package main

import (
	OrderInitiatorService "OrderInitiatorService/handlers"
	"Shared/network"
	"databaseAccessTransaction"
	"databaseAccessUserManagement"
	"fmt"
)

//"Shared/network"

func main() {
	//Need to upgrade to use my entity class stuff and the new services.

	networkManager := network.NewNetwork()

	databaseAccess := databaseAccessTransaction.NewDatabaseAccess(&databaseAccessTransaction.NewDatabaseAccessParams{
		Network: networkManager,
	})

	databaseAccessUserManagement := databaseAccessUserManagement.NewDatabaseAccess(&databaseAccessUserManagement.NewDatabaseAccessParams{
		Network: networkManager,
	})

	go OrderInitiatorService.InitalizeHandlers(networkManager, databaseAccess, databaseAccessUserManagement)
	fmt.Println("Matching Engine Service Started")

	networkManager.Listen(network.ListenerParams{
		Handler: nil,
	})
}
