package main

import (
	networkHttp "Shared/network/http"
	databaseServiceUserManagement "databaseServiceUserManagement/database-connection"
	userManagementDatabaseHandlers "databaseServiceUserManagement/handlers"
	"fmt"
)

//"Shared/network"

func main() {
	networkManager := networkHttp.NewNetworkHttp()
	_databaseManager := databaseServiceUserManagement.NewDatabaseService(&databaseServiceUserManagement.NewDatabaseServiceParams{})

	go userManagementDatabaseHandlers.InitalizeHandlers(networkManager, _databaseManager)
	fmt.Println("User Management Database Service Started")

	networkManager.Listen()
}
