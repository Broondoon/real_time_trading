package main

import (
	"Shared/network"
	databaseServiceUserManagement "databaseServiceUserManagement/database-connection"
	userManagementDatabaseHandlers "databaseServiceUserManagement/handlers"
	"fmt"
)

//"Shared/network"

func main() {
	networkManager := network.NewNetwork()
	_databaseManager := databaseServiceUserManagement.NewDatabaseService(&databaseServiceUserManagement.NewDatabaseServiceParams{})

	go userManagementDatabaseHandlers.InitalizeHandlers(networkManager, _databaseManager)
	fmt.Println("User Management Database Service Started")

	networkManager.Listen(network.ListenerParams{
		Handler: nil,
	})
}
