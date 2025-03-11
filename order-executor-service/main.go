package main

import (
	OrderExecutorService "OrderExecutorService/orderExecutor"
	networkHttp "Shared/network/http"
	networkQueue "Shared/network/queue"

	"databaseAccessTransaction"
	"databaseAccessUserManagement"
	"log"
	"os"
)

func main() {

	networkManager := networkHttp.NewNetworkHttp()
	networkQueueManager := networkQueue.NewNetworkQueue(nil, os.Getenv("ORDER_EXECUTOR_HOST")+":"+os.Getenv("ORDER_EXECUTOR_PORT"))

	databaseAccessTransaction := databaseAccessTransaction.NewDatabaseAccess(&databaseAccessTransaction.NewDatabaseAccessParams{
		Network: networkManager,
	})

	databaseAccessUserManagement := databaseAccessUserManagement.NewDatabaseAccess(&databaseAccessUserManagement.NewDatabaseAccessParams{
		Network: networkManager,
	})

	// Clarify what this is doing and why it is necessary
	go OrderExecutorService.InitalizeExecutorHandlers(networkManager, networkQueueManager, databaseAccessTransaction, databaseAccessUserManagement)
	log.Println("Order Executor Service Started")

	networkManager.Listen()

}
