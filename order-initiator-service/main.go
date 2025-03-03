package main

import (
	OrderInitiatorService "OrderInitiatorService/handlers"
	networkHttp "Shared/network/http"
	"databaseAccessTransaction"
	"databaseAccessUserManagement"
	"fmt"
)

//"Shared/network"

func main() {
	//Need to upgrade to use my entity class stuff and the new services.

	networkManager := networkHttp.NewNetworkHttp()

	databaseAccess := databaseAccessTransaction.NewDatabaseAccess(&databaseAccessTransaction.NewDatabaseAccessParams{
		Network: networkManager,
	})

	databaseAccessUserManagement := databaseAccessUserManagement.NewDatabaseAccess(&databaseAccessUserManagement.NewDatabaseAccessParams{
		Network: networkManager,
	})

	go OrderInitiatorService.InitalizeHandlers(networkManager, databaseAccess, databaseAccessUserManagement)
	fmt.Println("Matching Engine Service Started")

	networkManager.Listen()
}
