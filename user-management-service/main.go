package main

import (
	"Shared/network"
	"databaseAccessUserManagement"
	databaseServiceUserManagement "databaseServiceUserManagement/database-connection"
	"log"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"UserManagementService/handlers"
)

func main() {
	db, err := gorm.Open(sqlite.Open("user_management.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	dbService := databaseServiceUserManagement.NewDatabaseService(db)

	walletAccess := databaseAccessUserManagement.NewWalletDatabaseAccess(dbService)
	userStockAccess := databaseAccessUserManagement.NewUserStockDatabaseAccess(dbService)

	handlers.InitializeWallet(walletAccess)
	handlers.InitializeUserStock(userStockAccess)
	handlers.InitializeHealth()

	log.Println("User Management Service started on port", os.Getenv("USER_MANAGEMENT_PORT"))

	network.NewNetwork().Listen(network.ListenerParams{Handler: nil})
}
