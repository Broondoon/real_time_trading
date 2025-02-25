package handlers

import (
	userStock "Shared/entities/user-stock"
	"Shared/network"
	"databaseAccessUserManagement"
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
)

type AddStock struct {
	StockID  string `json:"stock_id"`
	Quantity int    `json:"quantity"`
}

var _userStockAccess databaseAccessUserManagement.UserStocksDataAccessInterface

func InitializeUserStock(userStockAccess databaseAccessUserManagement.UserStocksDataAccessInterface, networkManager network.NetworkInterface) {
	_userStockAccess = userStockAccess
	networkManager.AddHandleFuncProtected(network.HandlerParams{Pattern: "transaction/getStockPortfolio", Handler: getStockPortfolioHandler})
	networkManager.AddHandleFuncProtected(network.HandlerParams{Pattern: "setup/addStockToUser", Handler: addStockToUser})
}

func getStockPortfolioHandler(responseWriter http.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	userID := queryParams.Get("userID")
	if userID == "" {
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}

	stocks, err := _userStockAccess.GetUserStocks(userID)
	if err != nil {
		responseWriter.WriteHeader(http.StatusNotFound)
		return
	}
	returnVal := network.ReturnJSON{
		Success: true,
		Data:    stocks,
	}

	sort.SliceStable(*stocks, func(i, j int) bool {
		return (*stocks)[i].GetStockName() > (*stocks)[j].GetStockName()
	})
	returnVal := network.ReturnJSON{
		Success: true,
		Data:    stocks,
	}

	stocksJSON, err := json.Marshal(returnVal)
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	responseWriter.Write(stocksJSON)
}

func addStockToUser(responseWriter http.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	userID := queryParams.Get("userID")
	if userID == "" {
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}

	var stockRequest AddStock
	err := json.Unmarshal(data, &stockRequest)
	if err != nil {
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}

	if stockRequest.StockID == "" || stockRequest.Quantity <= 0 {
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}

	newUserStock := userStock.New(userStock.NewUserStockParams{
		UserID:    userID,
		StockID:   stockRequest.StockID,
		Quantity:  stockRequest.Quantity,
		StockName: "Unknown",
	})

	createdUserStock, err := _userStockAccess.Create(newUserStock)
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	returnVal := network.ReturnJSON{
		Success: true,
		Data:    createdUserStock,
	}

	responseJSON, err := json.Marshal(returnVal)
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusCreated)
	responseWriter.Write(responseJSON)
}
