package main

import (
	networkHttp "Shared/network/http"
	networkQueue "Shared/network/queue"
	databaseServiceStock "databaseServiceStock/database-connection"
	stockDatabaseHandlers "databaseServiceStock/handlers"
	"log"
	"os"
)

//"Shared/network"

func main() {
	networkManager := networkHttp.NewNetworkHttp()
	networkManagerQueue := networkQueue.NewNetworkQueue(nil, os.Getenv("STOCK_DATABASE_SERVICE_HOST")+":"+os.Getenv("STOCK_DATABASE_SERVICE_PORT"))

	_databaseManager := databaseServiceStock.NewDatabaseService(&databaseServiceStock.NewDatabaseServiceParams{})

	go stockDatabaseHandlers.InitalizeHandlers(networkManager, networkManagerQueue, _databaseManager)
	log.Println("Stock Database Service Started")

	networkManager.Listen()
	<-make(chan struct{})
}
