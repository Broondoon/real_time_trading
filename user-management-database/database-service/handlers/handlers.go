package userManagementDatabaseHandlers

import (
	userStock "Shared/entities/user-stock"
	"Shared/entities/wallet"
	"Shared/network"
	databaseServiceUserManagement "databaseServiceUserManagement/database-connection"
	"fmt"
	"net/http"
	"os"
)

var _databaseManager databaseServiceUserManagement.DatabaseServiceInterface
var _networkManager network.NetworkInterface

func InitalizeHandlers(
	networkManager network.NetworkInterface, databaseManager databaseServiceUserManagement.DatabaseServiceInterface) {
	_databaseManager = databaseManager
	_networkManager = networkManager

	//Add handlers
	network.CreateNetworkEntityHandlers[*userStock.UserStock](_networkManager, os.Getenv("USER_MANAGEMENT_SERVICE_USER_STOCK_ROUTE"), _databaseManager.UserStocks(), userStock.Parse)
	network.CreateNetworkEntityHandlers[*wallet.Wallet](_networkManager, os.Getenv("USER_MANAGEMENT_SERVICE_WALLET_ROUTE"), _databaseManager.Wallets(), wallet.Parse)
	http.HandleFunc("/health", healthHandler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Simple check: you might expand this to test database connectivity, etc.
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}
