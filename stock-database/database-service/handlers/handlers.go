package stockDatabaseHandlers

import (
	"Shared/entities/stock"
	"Shared/network"
	"databaseServiceStock/database-connection"
	"encoding/json"
	"net/http"
)

var _databaseManager databaseServiceStock.DatabaseServiceInterface
var _networkManager network.NetworkInterface

func InitalizeHandlers(
	networkManager network.NetworkInterface, databaseManager databaseServiceStock.DatabaseServiceInterface) {
	_databaseManager = databaseManager
	_networkManager = networkManager

	//Add handlers
	networkManager.AddHandleFunc(network.HandlerParams{Pattern: "/createStock", Handler: AddNewStockHandler})
	networkManager.AddHandleFunc(network.HandlerParams{Pattern: "/getStockIDs", Handler: GetStockIDsHandler})

}

func GetStockIDsHandler(responseWriter http.ResponseWriter, data []byte) {
	stocks, err := _databaseManager.GetAll()
	if err != nil {
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
// we're expecting {"StockID":"{id value}"}
func AddNewStockHandler(responseWriter http.ResponseWriter, data []byte) {
	newStock, err := stock.Parse(data)
	if err != nil {
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	err = _databaseManager.Create(newStock)
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	jsonData, err := newStock.EntityToJSON()
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	jsonData, err = _networkManager.MatchingEngine().Post("/createStock", jsonData)
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = responseWriter.Write(jsonData)
}
