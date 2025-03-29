package main

import (
	OrderInitiatorService "OrderInitiatorService/handlers"
	networkHttp "Shared/network/http"
	networkQueue "Shared/network/queue"
	"databaseAccessStock"
	"databaseAccessUserManagement"
	"log"
	"os"
)

//"Shared/network"

func main() {
	//Need to upgrade to use my entity class stuff and the new services.

	networkHttpManager := networkHttp.NewNetworkHttp()
	networkQueueManager := networkQueue.NewNetworkQueue(nil, os.Getenv("ORDER_INITIATOR_HOST")+":"+os.Getenv("ORDER_INITIATOR_PORT"))

	databaseAccessStock := databaseAccessStock.NewDatabaseAccess(&databaseAccessStock.NewDatabaseAccessParams{
		Network: networkQueueManager,
	})

	databaseAccessUserManagement := databaseAccessUserManagement.NewDatabaseAccess(&databaseAccessUserManagement.NewDatabaseAccessParams{
		Network: networkQueueManager,
	})

	go OrderInitiatorService.InitalizeHandlers(networkHttpManager, networkQueueManager, databaseAccessUserManagement, databaseAccessStock)
	log.Println("Matching Engine Service Started")

	networkHttpManager.Listen()
	<-make(chan struct{})
}
