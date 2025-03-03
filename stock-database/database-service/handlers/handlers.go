package stockDatabaseHandlers

import (
	"Shared/entities/stock"
	"Shared/network"
	databaseServiceStock "databaseServiceStock/database-connection"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
)

var _databaseManager databaseServiceStock.DatabaseServiceInterface
var _networkManager network.NetworkInterface

func InitalizeHandlers(
	networkManager network.NetworkInterface, databaseManager databaseServiceStock.DatabaseServiceInterface) {
	_databaseManager = databaseManager
	_networkManager = networkManager

	//Add handlers
	_networkManager.AddHandleFuncProtected(network.HandlerParams{Pattern: os.Getenv("setup_route") + "/createStock", Handler: AddNewStockHandler})
	_networkManager.AddHandleFuncUnprotected(network.HandlerParams{Pattern: "getStockIDs", Handler: GetStockIDsHandler})
	network.CreateNetworkEntityHandlers[*stock.Stock](_networkManager, os.Getenv("STOCK_DATABASE_SERVICE_ROUTE"), _databaseManager, stock.Parse)
	http.HandleFunc("/health", healthHandler)

}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Simple check: you might expand this to test database connectivity, etc.
	w.WriteHeader(http.StatusOK)
	////fmt.Println(w, "OK")
}

func GetStockIDsHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	stocks, err := _databaseManager.GetAll()
	if err != nil {
		println("Error: ", err.Error())
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	stockIDs := make([]string, len(*stocks))
	for i, stock := range *stocks {
		stockIDs[i] = stock.GetId()
	}
	stockIDsJSON, err := json.Marshal(stockIDs)
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	responseWriter.Write(stockIDsJSON)
}

// Expected input is a stock ID in the body of the request
func AddNewStockHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	newStock, err := stock.Parse(data)

	println("Parsed Stock: ", newStock.GetId())
	if err != nil {
		println("Error: ", err.Error())
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	err = _databaseManager.Create(newStock)
	println("Created Stock: ", newStock.GetId())
	if err != nil {
		println("Error: ", err.Error())
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	stockIdObject := network.StockID{StockID: newStock.GetId()}
	_, err = _networkManager.MatchingEngine().Post("createStock", stockIdObject)
	if err != nil {
		println("Error: ", err.Error())
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	returnVal := network.ReturnJSON{
		Success: true,
		Data:    network.StockID{StockID: newStock.GetId()},
	}
	returnValJSON, err := json.Marshal(returnVal)
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	responseWriter.Write(returnValJSON)
}
