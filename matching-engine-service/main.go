package main

import (
	"MatchingEngineService/matchingEngine"
	"Shared/entities/stock"
	"Shared/network"
	"os"
)

//"Shared/network"

func main() {
	stockList := []stock.StockInterface{}
	networkManager := network.NewHttpClient(os.Getenv("MATCHING_ENGINE_SERVICE_URL"))

	// Example usage of stockList to avoid "declared and not used" error
	if len(stockList) == 0 {
		println("No stocks available")
	}

	go matchingEngine.InitalizeHandlers(stockList, networkManager)

	//setup handlers:
	//placeStockOrder
	//listen for incoming order
	//create order object
	//if get order, create order object
	//if order is buy, add to buy order book
	//if order is sell, add to sell order book
	//cancelStockTransaction
	//listen for incoming order

}
