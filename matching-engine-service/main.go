package main

import (
	"MatchingEngineService/matchingEngine"
	"Shared/network"
	"databaseAccessStock"
	"databaseAccessStockOrder"
)

//"Shared/network"

func main() {
	//Need to upgrade to use my entity class stuff and the new services.

	networkManager := network.NewNetwork()
	_databaseManager := databaseAccessStockOrder.NewDatabaseAccess(&databaseAccessStockOrder.NewDatabaseAccessParams{})
	_databaseAccess := databaseAccessStock.NewDatabaseAccess(&databaseAccessStock.NewDatabaseAccessParams{
		Network: networkManager,
	})
	stockList, err := _databaseAccess.GetStockIDs()
	if err != nil {
		panic(err)
	}

	go matchingEngine.InitalizeHandlers(stockList, networkManager, _databaseManager, _databaseAccess)
	println("Matching Engine Service Started")

	networkManager.Listen(network.ListenerParams{
		Handler: nil,
	})
}
