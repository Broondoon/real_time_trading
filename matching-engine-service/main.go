package main

import (
	"MatchingEngineService/matchingEngine"
	"Shared/network"
	networkHttp "Shared/network/http"
	networkQueue "Shared/network/queue"
	"databaseAccessStock"
	"databaseAccessStockOrder"
	"databaseAccessUserManagement"
	"log"
	"os"
)

//"Shared/network"

func main() {
	//Need to upgrade to use my entity class stuff and the new services.

	networkHttpManager := networkHttp.NewNetworkHttp()
	networkQueueManager := networkQueue.NewNetworkQueue(nil, os.Getenv("MATCHING_ENGINE_HOST")+":"+os.Getenv("MATCHING_ENGINE_PORT"))
	_databaseManagerStockOrder := databaseAccessStockOrder.NewDatabaseAccess(&databaseAccessStockOrder.NewDatabaseAccessParams{})
	_databaseAccessStock := databaseAccessStock.NewDatabaseAccess(&databaseAccessStock.NewDatabaseAccessParams{
		Network: networkQueueManager,
	})
	_databaseAccessUserManagement := databaseAccessUserManagement.NewDatabaseAccess(&databaseAccessUserManagement.NewDatabaseAccessParams{
		Network: networkQueueManager,
	})
	stockList, err := _databaseAccessStock.GetAll()
	if err != nil {
		panic(err)
	}
	var stockList2 []network.StockPrice
	for _, stock := range *stockList {
		stockList2 = append(stockList2, network.StockPrice{
			StockID:   stock.GetIdString(),
			StockName: stock.GetName(),
		})
	}

	go matchingEngine.InitalizeHandlers(&stockList2, networkHttpManager, networkQueueManager, _databaseManagerStockOrder, _databaseAccessUserManagement)
	log.Println("Matching Engine Service Started")

	networkHttpManager.Listen()
	<-make(chan struct{})
}
