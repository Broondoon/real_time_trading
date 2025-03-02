package authDatabaseHandlers

import (
	"Shared/entities/user"
	"Shared/network"
	databaseServiceAuth "databaseServiceAuth/database-connection"
	"os"
)

var _databaseManager databaseServiceAuth.databaseServiceInterface
var _networkManager network.NetworkInterface

func InitializeHandlers(
	networkManager network.NetworkInterface, databaseManager databaseServiceAuth.databaseServiceInterface) {
	_databaseManager = databaseManager
	_networkManager = networkManager

	network.CreateNetworkEntityHandlers(_networkManager, os.Getenv("AUTH_SERVICE_USER_ROUTE"), _databaseManager.GetUsers(), user.Parse)
}
