package main

import (
	OrderExecutorService "OrderExecutorService/orderExecutor"
	networkHttp "Shared/network/http"
	networkQueue "Shared/network/queue"

	"databaseAccessUserManagement"
	"log"
	"os"
)

func main() {

	networkManager := networkHttp.NewNetworkHttp()
	networkQueueManager := networkQueue.NewNetworkQueue(nil, os.Getenv("ORDER_EXECUTOR_HOST")+":"+os.Getenv("ORDER_EXECUTOR_PORT"))

	databaseAccessUserManagement := databaseAccessUserManagement.NewDatabaseAccess(&databaseAccessUserManagement.NewDatabaseAccessParams{
		Network: networkManager,
	})

	// Clarify what this is doing and why it is necessary
	go OrderExecutorService.InitalizeExecutorHandlers(networkManager, networkQueueManager, databaseAccessUserManagement)
	log.Println("Order Executor Service Started")

	networkManager.Listen()
	<-make(chan struct{})
}
