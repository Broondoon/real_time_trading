package main

import (
	networkHttp "Shared/network/http"
	networkQueue "Shared/network/queue"
	databaseServiceUserManagement "databaseServiceUserManagement/database-connection"
	userManagementDatabaseHandlers "databaseServiceUserManagement/handlers"
	"log"
	"os"
)

//"Shared/network"

func main() {
	networkManagerHTTP := networkHttp.NewNetworkHttp()
	networkManagerQueue := networkQueue.NewNetworkQueue(nil, os.Getenv("USER_MANAGEMENT_DATABASE_SERVICE_HOST")+":"+os.Getenv("USER_MANAGEMENT_DATABASE_SERVICE_PORT"))
	_databaseManager := databaseServiceUserManagement.NewDatabaseService(&databaseServiceUserManagement.NewDatabaseServiceParams{})

	go userManagementDatabaseHandlers.InitalizeHandlers(networkManagerHTTP, networkManagerQueue, _databaseManager)
	log.Println("User Management Database Service Started")

	networkManagerHTTP.Listen()
	<-make(chan struct{})
}
