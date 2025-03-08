package main

import (
	networkHttp "Shared/network/http"
	databaseServiceTransaction "databaseServiceTransaction/database-connection"
	transactionDatabaseHandlers "databaseServiceTransaction/handlers"
	"fmt"
)

//"Shared/network"

func main() {
	networkManager := networkHttp.NewNetworkHttp()
	_databaseManager := databaseServiceTransaction.NewDatabaseService(&databaseServiceTransaction.NewDatabaseServiceParams{})

	go transactionDatabaseHandlers.InitalizeHandlers(networkManager, _databaseManager)
	fmt.Println("Transaction Database Service Started")

	networkManager.Listen()
}
