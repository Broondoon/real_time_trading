package handlers

import (
	userStock "Shared/entities/user-stock"
	"Shared/network"
	"databaseAccessUserManagement"
	"encoding/json"
	"log"
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
	//TODO:
	//testFuncInsertUserStock("6fd2fc6b-9142-4777-8b30-575ff6fa2460")
}

// TODO: delete this is for testing
/*
func testFuncInsertUserStock(userID string) {
	stockID1 := uuid.New().String()
	stockID2 := uuid.New().String()

	createdUserStock1, err := _userStockAccess.Create(userStock.New(userStock.NewUserStockParams{
		NewEntityParams: entity.NewEntityParams{
			ID: uuid.New().String(),
		},
		UserID:    userID,
		StockID:   stockID1,
		Quantity:  100,
		StockName: "AAPL",
	}))
	if err != nil {
		log.Fatalf("Failed to create user stock 1: %v", err)
	}
	fmt.Printf("Created user stock for user %s with stockID %s and quantity %d\n",
		createdUserStock1.GetUserID(), stockID1, createdUserStock1.GetQuantity())

	createdUserStock2, err := _userStockAccess.Create(userStock.New(userStock.NewUserStockParams{
		NewEntityParams: entity.NewEntityParams{
			ID: uuid.New().String(),
		},
		UserID:    userID,
		StockID:   stockID2,
		Quantity:  100,
		StockName: "GOOGL",
	}))
	if err != nil {
		log.Fatalf("Failed to create user stock 2: %v", err)
	}
	fmt.Printf("Created user stock for user %s with stockID %s and quantity %d\n",
		createdUserStock2.GetUserID(), stockID2, createdUserStock2.GetQuantity())
}
*/

func getStockPortfolioHandler(responseWriter http.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	userID := queryParams.Get("userID")
	if userID == "" {
		log.Println("Error: missing userID in getStockPortfolioHandler")
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	stocks, err := _userStockAccess.GetUserStocks(userID)
	if err != nil {
		log.Printf("Error retrieving stocks for userID %s: %v", userID, err)
		responseWriter.WriteHeader(http.StatusNotFound)
		return
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
		log.Printf("Error marshalling JSON in getStockPortfolioHandler for userID %s: %v", userID, err)
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	responseWriter.Write(stocksJSON)
}

func addStockToUser(responseWriter http.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	log.Printf("DEBUG: addStockToUser invoked. Request Type: %s, Query Params: %v, Request Body: %s", requestType, queryParams, string(data))

	userID := queryParams.Get("userID")
	if userID == "" {
		log.Println("ERROR: Missing userID in addStockToUser")
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("DEBUG: Extracted userID: %s", userID)

	var stockRequest AddStock
	err := json.Unmarshal(data, &stockRequest)
	if err != nil {
		log.Printf("ERROR: Failed to unmarshal request data in addStockToUser: %v", err)
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("DEBUG: Parsed AddStock request: %+v", stockRequest)

	if stockRequest.StockID == "" || stockRequest.Quantity <= 0 {
		log.Println("ERROR: Invalid stockRequest values in addStockToUser. StockID is empty or Quantity is non-positive.")
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}

	newUserStock := userStock.New(userStock.NewUserStockParams{
		UserID:    userID,
		StockID:   stockRequest.StockID,
		Quantity:  stockRequest.Quantity,
		StockName: "Unknown",
	})
	log.Printf("DEBUG: Created newUserStock object: %+v", newUserStock)

	createdUserStock, err := _userStockAccess.Create(newUserStock)
	if err != nil {
		log.Printf("ERROR: Failed to create user stock for userID %s: %v", userID, err)
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("DEBUG: Successfully created user stock: %+v", createdUserStock)

	returnVal := network.ReturnJSON{
		Success: true,
		Data:    nil,
	}
	responseJSON, err := json.Marshal(returnVal)
	if err != nil {
		log.Printf("ERROR: Failed to marshal response JSON in addStockToUser for userID %s: %v", userID, err)
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("DEBUG: Marshalled response JSON: %s", string(responseJSON))

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusCreated)
	_, err = responseWriter.Write(responseJSON)
	if err != nil {
		log.Printf("ERROR: Failed to write response for userID %s: %v", userID, err)
	}
	log.Println("DEBUG: addStockToUser completed successfully.")
}
