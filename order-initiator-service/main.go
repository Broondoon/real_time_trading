package main

import (
	OrderInitiatorService "OrderInitiatorService/handlers"
	networkHttp "Shared/network/http"
	networkQueue "Shared/network/queue"
	"databaseAccessTransaction"
	"databaseAccessUserManagement"
	"fmt"
	"os"
)

//"Shared/network"

func main() {
	//Need to upgrade to use my entity class stuff and the new services.

	networkHttpManager := networkHttp.NewNetworkHttp()
	networkQueueManager := networkQueue.NewNetworkQueue(nil, os.Getenv("ORDER_INITIATOR_HOST")+":"+os.Getenv("ORDER_INITIATOR_PORT"))

	databaseAccess := databaseAccessTransaction.NewDatabaseAccess(&databaseAccessTransaction.NewDatabaseAccessParams{
		Network: networkHttpManager,
	})

	databaseAccessUserManagement := databaseAccessUserManagement.NewDatabaseAccess(&databaseAccessUserManagement.NewDatabaseAccessParams{
		Network: networkHttpManager,
	})

	go OrderInitiatorService.InitalizeHandlers(networkHttpManager, networkQueueManager, databaseAccess, databaseAccessUserManagement)
	fmt.Println("Matching Engine Service Started")

	networkHttpManager.Listen()
	<-make(chan struct{})
}
