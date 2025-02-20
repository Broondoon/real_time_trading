package handlers

import (
	"databaseAccessUserManagement"
	"encoding/json"
	"net/http"
)

var _walletAccess databaseAccessUserManagement.WalletDataAccessInterface

func InitializeWallet(walletAccess databaseAccessUserManagement.WalletDataAccessInterface) {
	_walletAccess = walletAccess

	http.HandleFunc("/getWalletBalance", getWalletBalanceHandler)
	http.HandleFunc("/addMoneyToWallet", addMoneyToWalletHandler)
}

func getWalletBalanceHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		http.Error(w, "Missing userID", http.StatusBadRequest)
		return
	}

	balance, err := _walletAccess.GetWalletBalance(userID)
	if err != nil {
		http.Error(w, "Failed to get wallet balance", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]float64{"balance": balance})
}

func addMoneyToWalletHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		UserID string  `json:"userID"`
		Amount float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := _walletAccess.AddMoneyToWallet(request.UserID, request.Amount)
	if err != nil {
		http.Error(w, "Failed to add money to wallet", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
