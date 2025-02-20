package stockDatabaseHandlers

import (
	"Shared/entities/stock"
	"Shared/network"
	databaseServiceStock "databaseServiceStock/database-connection"
	"encoding/json"
	"fmt"
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
	_networkManager.AddHandleFunc(network.HandlerParams{Pattern: "createStock", Handler: AddNewStockHandler})
	_networkManager.AddHandleFunc(network.HandlerParams{Pattern: "getStockIDs", Handler: GetStockIDsHandler})
	network.CreateNetworkEntityHandlers[*stock.Stock](_networkManager, os.Getenv("STOCK_DATABASE_SERVICE_ROUTE"), _databaseManager, stock.Parse)
	http.HandleFunc("/health", healthHandler)

}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Simple check: you might expand this to test database connectivity, etc.
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func GetStockIDsHandler(responseWriter http.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
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
func AddNewStockHandler(responseWriter http.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
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
