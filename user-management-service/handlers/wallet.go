package handlers

import (
	"Shared/entities/entity"
	"Shared/entities/wallet"
	"Shared/network"
	"databaseAccessUserManagement"
	"encoding/json"
	"fmt"
	"log"
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
	networkManager.AddHandleFuncProtected(network.HandlerParams{Pattern: "transaction/addMoneyToWallet", Handler: addMoneyToWalletHandler})
	//TODO: Comment out below line when not testing:
	testFuncInsertIntoDb("6fd2fc6b-9142-4777-8b30-575ff6fa2460")

}

func getWalletBalanceHandler(responseWriter http.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	userID := queryParams.Get("userID")
	if userID == "" {
		// Fallback if "userID" isn’t provided.
		userID = queryParams.Get("id")
	}
	fmt.Printf("Received request to get wallet balance. userID=%s\n", userID)

	fmt.Printf("Request Type: %s\n", requestType)
	fmt.Printf("Query Params: %v\n", queryParams)
	fmt.Printf("Request Body: %s\n", string(data))
	fmt.Printf("Extracted userID: %s\n", userID)

	if userID == "" {
		fmt.Println("Error: Missing userID in query parameters.")
		responseWriter.WriteHeader(http.StatusBadRequest)
		fmt.Println("===== [END] getWalletBalanceHandler - Failed: Missing userID =====")
		return
	}

	balance, err := _walletAccess.GetWalletBalance(userID)
	if err != nil {
		fmt.Printf("Error: Failed to get wallet balance for userID=%s. Reason: %v\n", userID, err)
		responseWriter.WriteHeader(http.StatusBadRequest)
		fmt.Println("===== [END] getWalletBalanceHandler - Failed: Database Error =====")
		return
	}

	walletBalance := WalletBalance{Balance: balance}

	walletJSON, err := json.Marshal(walletBalance)
	if err != nil {
		http.Error(responseWriter, "Failed to marshal wallet balance: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Sending successful response...")
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusOK)
	responseWriter.Write(walletJSON)
}

// TODO: comment this out later
func testFuncInsertIntoDb(userID string) {
	params := wallet.NewWalletParams{
		NewEntityParams: entity.NewEntityParams{},
		UserID:          userID,
		Balance:         100.0,
	}
	newWallet := wallet.New(params)

	createdWallet, err := _walletAccess.Create(newWallet)
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}
	fmt.Printf("Created wallet for user %s with balance: %.2f\n", createdWallet.GetUserID(), createdWallet.GetBalance())
}

func addMoneyToWalletHandler(responseWriter http.ResponseWriter, data []byte, queryParams url.Values, requestType string) {

	fmt.Printf("DEBUG: Received addMoneyToWallet request. Request Type: %s, Query Params: %v\n", requestType, queryParams)

	userID := queryParams.Get("userID")
	fmt.Printf("DEBUG: Extracted userID: %s\n", userID)
	if userID == "" {
		fmt.Println("DEBUG: userID is missing, returning 400 Bad Request")
		responseWriter.WriteHeader(http.StatusBadRequest)
		responseWriter.Write([]byte("Missing userID"))
		return
	}

	fmt.Printf("DEBUG: Raw request data: %s\n", string(data))
	var request struct {
		Amount float64 `json:"amount"`
	}

	if err := json.Unmarshal(data, &request); err != nil {
		fmt.Println("DEBUG: Error unmarshalling request data:", err.Error())
		responseWriter.WriteHeader(http.StatusBadRequest)
		responseWriter.Write([]byte("Invalid request body"))
		return
	}
	fmt.Printf("DEBUG: Parsed request amount: %f\n", request.Amount)

	if request.Amount <= 0 {
		fmt.Println("DEBUG: Request amount is invalid (<= 0), returning 400 Bad Request")
		responseWriter.WriteHeader(http.StatusBadRequest)
		responseWriter.Write([]byte("Amount must be greater than zero"))
		return
	}

	fmt.Printf("DEBUG: Calling _walletAccess.AddMoneyToWallet for userID %s with amount %f\n", userID, request.Amount)
	if err := _walletAccess.AddMoneyToWallet(userID, request.Amount); err != nil {
		fmt.Printf("DEBUG: Error adding money to wallet: %v\n", err)
		responseWriter.WriteHeader(http.StatusInternalServerError)
		responseWriter.Write([]byte("Failed to add money to wallet"))
		return
	}

	fmt.Println("DEBUG: Money added successfully, sending 200 OK response")
	responseWriter.WriteHeader(http.StatusOK)
	responseWriter.Write([]byte("Money added successfully"))
}
