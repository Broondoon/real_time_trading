package handlers

import (
	"databaseAccessUserManagement"
	"encoding/json"
	"net/http"
)

var _userStockAccess databaseAccessUserManagement.UserStockDataAccessInterface

func InitializeUserStock(userStockAccess databaseAccessUserManagement.UserStockDataAccessInterface) {
	_userStockAccess = userStockAccess

	http.HandleFunc("/getStockPortfolio", getStockPortfolioHandler)
}

func getStockPortfolioHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		http.Error(w, "Missing userID", http.StatusBadRequest)
		return
	}

	stocks, err := _userStockAccess.GetUserStocks(userID)
	if err != nil {
		http.Error(w, "Failed to get stock portfolio", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stocks)
}
