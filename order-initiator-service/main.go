package main

import (
	OrderInitiatorService "OrderInitiatorService/handlers"
	networkHttp "Shared/network/http"
	networkQueue "Shared/network/queue"
	"databaseAccessStock"
	"databaseAccessTransaction"
	"databaseAccessUserManagement"
	"log"
	"os"
)

//"Shared/network"

func main() {
	//Need to upgrade to use my entity class stuff and the new services.

	networkHttpManager := networkHttp.NewNetworkHttp()
	networkQueueManager := networkQueue.NewNetworkQueue(nil, os.Getenv("ORDER_INITIATOR_HOST")+":"+os.Getenv("ORDER_INITIATOR_PORT"))

	databaseAccessTransaction := databaseAccessTransaction.NewDatabaseAccess(&databaseAccessTransaction.NewDatabaseAccessParams{
		Network: networkHttpManager,
	})

	databaseAccessStock := databaseAccessStock.NewDatabaseAccess(&databaseAccessStock.NewDatabaseAccessParams{
		Network: networkHttpManager,
	})

	databaseAccessUserManagement := databaseAccessUserManagement.NewDatabaseAccess(&databaseAccessUserManagement.NewDatabaseAccessParams{
		Network: networkHttpManager,
	})

	go OrderInitiatorService.InitalizeHandlers(networkHttpManager, networkQueueManager, databaseAccessTransaction, databaseAccessUserManagement, databaseAccessStock)
	log.Println("Matching Engine Service Started")

	networkHttpManager.Listen()
	<-make(chan struct{})
}
