package handlers

import (
	"Shared/network"
	"databaseAccessUserManagement"
	"encoding/json"
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
	//networkManager.AddHandleFuncUnprotected(network.HandlerParams{Pattern: "transaction/createWallet", Handler: createWalletHandler})

	//TODO: Comment out below line when not testing:
	//testFuncInsertIntoDb("6fd2fc6b-9142-4777-8b30-575ff6fa2460")

}

// TODO: comment this out later
/*func testFuncInsertIntoDb(userID string) {
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
	log.Printf("Created wallet for user %s with balance: %.2f\n", createdWallet.GetUserID(), createdWallet.GetBalance())
}*/

func getWalletBalanceHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	userID := queryParams.Get("userID")
	if userID == "" {
		// Fallback if "userID" isn’t provided.
		userID = queryParams.Get("id")
	}
	log.Printf("Received request to get wallet balance. userID=%s\n", userID)

	log.Printf("Request Type: %s\n", requestType)
	log.Printf("Query Params: %v\n", queryParams)
	log.Printf("Request Body: %s\n", string(data))
	log.Printf("Extracted userID: %s\n", userID)

	if userID == "" {
		log.Println("Error: Missing userID in query parameters.")
		responseWriter.WriteHeader(http.StatusBadRequest)
		log.Println("===== [END] getWalletBalanceHandler - Failed: Missing userID =====")
		return
	}

	balance, err := _walletAccess.GetWalletBalance(userID)
	if err != nil {
		log.Printf("Error: Failed to get wallet balance for userID=%s. Reason: %v\n", userID, err)
		responseWriter.WriteHeader(http.StatusBadRequest)
		log.Println("===== [END] getWalletBalanceHandler - Failed: Database Error =====")
		return
	}

	walletBalance := WalletBalance{Balance: balance}
	returnVal := network.ReturnJSON{
		Success: true,
		Data:    walletBalance,
	}

	walletJSON, err := json.Marshal(returnVal)
	if err != nil {
		defer responseWriter.WriteHeader(http.StatusInternalServerError)
		//http.Error(responseWriter, "Failed to marshal wallet balance: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Sending successful response...")
	responseWriter.Write(walletJSON)
}

func addMoneyToWalletHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {

	log.Printf("DEBUG: Received addMoneyToWallet request. Request Type: %s, Query Params: %v\n", requestType, queryParams)

	userID := queryParams.Get("userID")
	log.Printf("DEBUG: Extracted userID: %s\n", userID)
	if userID == "" {
		log.Println("DEBUG: userID is missing, returning 400 Bad Request")
		responseWriter.WriteHeader(http.StatusBadRequest)
		responseWriter.Write([]byte("Missing userID"))
		return
	}

	log.Printf("DEBUG: Raw request data: %s\n", string(data))
	var request struct {
		Amount float64 `json:"amount"`
	}

	if err := json.Unmarshal(data, &request); err != nil {
		log.Println("DEBUG: Error unmarshalling request data:", err.Error())
		responseWriter.WriteHeader(http.StatusBadRequest)
		responseWriter.Write([]byte("Invalid request body"))
		return
	}
	log.Printf("DEBUG: Parsed request amount: %f\n", request.Amount)

	if request.Amount <= 0 {
		log.Println("DEBUG: Request amount is invalid (<= 0), returning 400 Bad Request")
		responseWriter.WriteHeader(http.StatusBadRequest)
		responseWriter.Write([]byte("Amount must be greater than zero"))
		return
	}

	log.Printf("DEBUG: Calling _walletAccess.AddMoneyToWallet for userID %s with amount %f\n", userID, request.Amount)
	if err := _walletAccess.AddMoneyToWallet(userID, request.Amount); err != nil {
		log.Printf("DEBUG: Error adding money to wallet: %v\n", err)
		responseWriter.WriteHeader(http.StatusInternalServerError)
		responseWriter.Write([]byte("Failed to add money to wallet"))
		return
	}
	returnVal := network.ReturnJSON{
		Success: true,
		Data:    nil,
	}

	returnValJSON, err := json.Marshal(returnVal)
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	responseWriter.Write(returnValJSON)

	log.Println("DEBUG: Money added successfully, sending 200 OK response")
	responseWriter.WriteHeader(http.StatusOK)
}

// func createWalletHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
// 	log.Printf("DEBUG: createWalletHandler invoked. Request Type: %s, Query Params: %v, Request Body: %s\n", requestType, queryParams, string(data))

// 	userID := queryParams.Get("userID")
// 	if userID == "" {
// 		log.Println("DEBUG: Missing userID in query parameters.")
// 		responseWriter.WriteHeader(http.StatusBadRequest)
// 		return
// 	}
// 	log.Printf("DEBUG: Extracted userID: %s\n", userID)

// 	params := wallet.NewWalletParams{
// 		NewEntityParams: entity.NewEntityParams{},
// 		UserID:          userID,
// 		Balance:         0.0,
// 	}
// 	newWallet := wallet.New(params)
// 	log.Printf("DEBUG: Created wallet object for userID: %s\n", userID)

// 	createdWallet, err := _walletAccess.Create(newWallet)
// 	if err != nil {
// 		log.Printf("DEBUG: Failed to create wallet for userID: %s, error: %v\n", userID, err)
// 		responseWriter.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}
// 	log.Printf("DEBUG: Successfully created wallet for userID: %s. Wallet details: %+v\n", userID, createdWallet)

// 	responseWriter.WriteHeader(http.StatusOK)
// 	log.Println("DEBUG: createWalletHandler completed successfully.")
// }
