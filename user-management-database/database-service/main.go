package main

import (
	"Shared/network"
	databaseServiceUserManagement "databaseServiceUserManagement/database-connection"
	userManagementDatabaseHandlers "databaseServiceUserManagement/handlers"
)

//"Shared/network"

func main() {
	networkManager := network.NewNetwork()
	_databaseManager := databaseServiceUserManagement.NewDatabaseService(&databaseServiceUserManagement.NewDatabaseServiceParams{})

	go userManagementDatabaseHandlers.InitalizeHandlers(networkManager, _databaseManager)
	println("User Management Database Service Started")

	networkManager.Listen(network.ListenerParams{
		Handler: nil,
	})
}
