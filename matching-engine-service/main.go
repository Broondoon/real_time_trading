package main

import (
	"MatchingEngineService/matchingEngine"
	"Shared/network"
	"databaseAccessStockOrder"
)

//"Shared/network"

func main() {
	//Need to upgrade to use my entity class stuff and the new services.

	networkManager := network.NewNetwork()

	// body, err := networkManager.UserManagement().Get("/getStockIDs", nil) //probably not the right service to be calling, but I'm not curretnly sure whose storing stocks
	// if err != nil {
	// 	println("Error getting stock IDs")
	// 	return
	// }
	// var stockList []string

	// var jsonData map[string]interface{}
	// errJson := json.Unmarshal(body, &jsonData)
	// if errJson != nil {
	// 	return
	// }
	// stockList = jsonData["StockIDs"].([]string)
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
