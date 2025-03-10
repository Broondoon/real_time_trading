package main

import (
	networkHttp "Shared/network/http"
	databaseServiceStock "databaseServiceStock/database-connection"
	stockDatabaseHandlers "databaseServiceStock/handlers"
	"log"
)

//"Shared/network"

func main() {
	networkManager := networkHttp.NewNetworkHttp()
	_databaseManager := databaseServiceStock.NewDatabaseService(&databaseServiceStock.NewDatabaseServiceParams{})

	go stockDatabaseHandlers.InitalizeHandlers(networkManager, _databaseManager)
	log.Println("Stock Database Service Started")

	networkManager.Listen()
}
