package main

import (
	"MatchingEngineService/matchingEngine"
	"Shared/network"
	"databaseAccessStockOrder"
	"encoding/json"
	"os"
)

//"Shared/network"

func main() {
	networkManager := network.NewHttpClient(os.Getenv("MATCHING_ENGINE_SERVICE_URL"))
	body, err := networkManager.Get("/getStockIDs", nil)
	if err != nil {
		println("Error getting stock IDs")
		return
	}
	var stockList []string

	var jsonData map[string]interface{}
	errJson := json.Unmarshal(body, &jsonData)
	if errJson != nil {
		return
	}
	stockList = jsonData["StockIDs"].([]string)
	_databaseManager := databaseAccessStockOrder.NewDatabaseAccess(&databaseAccessStockOrder.NewDatabaseAccessParams{})

	// Example usage of stockList to avoid "declared and not used" error
	if len(stockList) == 0 {
		println("No stocks available")
	}

	go matchingEngine.InitalizeHandlers(&stockList, networkManager, _databaseManager)
	println("Matching Engine Service Started")

	networkManager.Listen(network.ListenerParams{
		Port:    "8080",
		Handler: nil,
	})
}
