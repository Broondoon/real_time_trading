package main

import (
	"MatchingEngineService/matchingEngine"
	networkHttp "Shared/network/http"
	networkQueue "Shared/network/queue"
	"databaseAccessStock"
	"databaseAccessStockOrder"
	"fmt"
	"os"
)

//"Shared/network"

func main() {
	//Need to upgrade to use my entity class stuff and the new services.

	networkHttpManager := networkHttp.NewNetworkHttp()
	networkQueueManager := networkQueue.NewNetworkQueue(nil, os.Getenv("MATCHING_ENGINE_HOST")+":"+os.Getenv("MATCHING_ENGINE_PORT"))
	_databaseManager := databaseAccessStockOrder.NewDatabaseAccess(&databaseAccessStockOrder.NewDatabaseAccessParams{})
	_databaseAccess := databaseAccessStock.NewDatabaseAccess(&databaseAccessStock.NewDatabaseAccessParams{
		Network: networkHttpManager,
	})
	stockList, err := _databaseAccess.GetStockIDs()
	if err != nil {
		panic(err)
	}

	go matchingEngine.InitalizeHandlers(stockList, networkHttpManager, networkQueueManager, _databaseManager, _databaseAccess)
	fmt.Println("Matching Engine Service Started")

	networkHttpManager.Listen()
}
