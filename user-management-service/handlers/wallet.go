package handlers

import (
	"Shared/network"
	"databaseAccessUserManagement"
	"encoding/json"
	"net/http"
	"net/url"
)

type WalletBalance struct {
	Balance float64 `json:"balance"`
}

var _walletAccess databaseAccessUserManagement.WalletDataAccessInterface

func InitializeWallet(walletAccess databaseAccessUserManagement.WalletDataAccessInterface, networkManager network.NetworkInterface) {
	_walletAccess = walletAccess

	networkManager.AddHandleFuncProtected(network.HandlerParams{Pattern: "transaction/getWalletBalance", Handler: getWalletBalanceHandler})
	networkManager.AddHandleFuncProtected(network.HandlerParams{Pattern: "transaction/addMoneyToWallet", Handler: getWalletBalanceHandler})
}

func getWalletBalanceHandler(responseWriter http.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	userID := queryParams.Get("userID")

	if userID == "" {
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}

	balance, err := _walletAccess.GetWalletBalance(userID)
	if err != nil {
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}

	walletBalance := WalletBalance{Balance: balance}

	walletJSON, err := json.Marshal(walletBalance)
	if err != nil {
		http.Error(responseWriter, "Failed to marshal wallet balance: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusOK)
	responseWriter.Write(walletJSON)
}

func addMoneyToWalletHandler(responseWriter http.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	userID := queryParams.Get("userID")
	if userID == "" {
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}

	var request struct {
		Amount float64 `json:"amount"`
	}

	if err := json.Unmarshal(data, &request); err != nil {
		responseWriter.WriteHeader(http.StatusBadRequest)
		responseWriter.Write([]byte("Invalid request body"))
		return
	}

	if request.Amount <= 0 {
		responseWriter.WriteHeader(http.StatusBadRequest)
		responseWriter.Write([]byte("Amount must be greater than zero"))
		return
	}

	if err := _walletAccess.AddMoneyToWallet(userID, request.Amount); err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		responseWriter.Write([]byte("Failed to add money to wallet"))
		return
	}

	responseWriter.WriteHeader(http.StatusOK)
	responseWriter.Write([]byte("Money added successfully"))
}
