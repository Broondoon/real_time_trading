package main

import (
	OrderExecutorService "OrderExecutorService/orderExecutor"
	networkHttp "Shared/network/http"
	"databaseAccessTransaction"
	"databaseAccessUserManagement"
)

func main() {

	networkManager := networkHttp.NewNetworkHttp()

	databaseAccessTransaction := databaseAccessTransaction.NewDatabaseAccess(&databaseAccessTransaction.NewDatabaseAccessParams{
		Network: networkManager,
	})

	databaseAccessUserManagement := databaseAccessUserManagement.NewDatabaseAccess(&databaseAccessUserManagement.NewDatabaseAccessParams{
		Network: networkManager,
	})

	// Clarify what this is doing and why it is necessary
	go OrderExecutorService.InitalizeExecutorHandlers(networkManager, databaseAccessTransaction, databaseAccessUserManagement)
	println("Order Executor Service Started")

	networkManager.Listen()

}
