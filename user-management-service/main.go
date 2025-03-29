package main

import (
	networkHttp "Shared/network/http"
	networkQueue "Shared/network/queue"
	"databaseAccessStock"
	"databaseAccessUserManagement"
	"log"
	"os"

	"UserManagementService/handlers"
)

func main() {

	networkManager := networkHttp.NewNetworkHttp()
	networkManagerQueue := networkQueue.NewNetworkQueue(nil, os.Getenv("USER_MANAGEMENT_HOST")+":"+os.Getenv("USER_MANAGEMENT_PORT"))

	databaseAccess := databaseAccessUserManagement.NewDatabaseAccess(&databaseAccessUserManagement.NewDatabaseAccessParams{
		Network: networkManagerQueue,
	})
	stockDatabaseAccess := databaseAccessStock.NewDatabaseAccess(&databaseAccessStock.NewDatabaseAccessParams{
		Network: networkManagerQueue,
	})

	walletAccess := databaseAccess.Wallet()
	userStockAccess := databaseAccess.UserStock()

	handlers.InitializeWallet(walletAccess, networkManager)
	handlers.InitializeUserStock(userStockAccess, stockDatabaseAccess, networkManager)
	handlers.InitializeHealth()

	log.Println("User Management Service started on port", os.Getenv("USER_MANAGEMENT_PORT"))

	networkManager.Listen()
	<-make(chan struct{})
}
