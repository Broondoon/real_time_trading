package authDatabaseHandlers

import (
	"net/http"
	"os"

	"Shared/entities/user"
	"Shared/network"
	databaseServiceAuth "databaseServiceAuth/database-connection"
)

var _databaseManager databaseServiceAuth.DatabaseServiceInterface
var _networkManager network.NetworkInterface

func InitializeHandlers(
	networkManager network.NetworkInterface, databaseManager databaseServiceAuth.DatabaseServiceInterface) {

	_databaseManager = databaseManager
	_networkManager = networkManager

	// Add handlers here
	network.CreateNetworkEntityHandlers(_networkManager, os.Getenv("AUTH_SERVICE_USER_ROUTE"), _databaseManager.User(), user.Parse)
	http.HandleFunc("/health", healthHandler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Simple check: you might expand this to test database connectivity, etc.
	w.WriteHeader(http.StatusOK)
	//fmt.Println(w, "OK")
}
