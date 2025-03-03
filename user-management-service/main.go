package main

import (
	"Shared/network"
	"databaseAccessUserManagement"
	"databaseAccessStock"
	"log"
	"os"

	"UserManagementService/handlers"
)

func main() {

	networkManager := network.NewNetwork()
	databaseAccess := databaseAccessUserManagement.NewDatabaseAccess(&databaseAccessUserManagement.NewDatabaseAccessParams{
		Network: networkManager,
	})
	stockDatabaseAccess := databaseAccessStock.NewDatabaseAccess(&databaseAccessStock.NewDatabaseAccessParams{
		Network: networkManager,
	})

	walletAccess := databaseAccess.Wallet()
	userStockAccess := databaseAccess.UserStock()


	handlers.InitializeWallet(walletAccess, networkManager)
	handlers.InitializeUserStock(userStockAccess, stockDatabaseAccess, networkManager)
	handlers.InitializeHealth()

	log.Println("User Management Service started on port", os.Getenv("USER_MANAGEMENT_PORT"))

	networkManager.Listen(network.ListenerParams{Handler: nil})
}
