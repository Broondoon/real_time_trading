package handlers

import (
	userStock "Shared/entities/user-stock"
	"Shared/network"
	"databaseAccessStock"
	"databaseAccessUserManagement"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"time"
)

type AddStock struct {
	StockID  string `json:"stock_id"`
	Quantity int    `json:"quantity"`
}

var _userStockAccess databaseAccessUserManagement.UserStocksDataAccessInterface
var _stockDatabaseAccess databaseAccessStock.DatabaseAccessInterface

func InitializeUserStock(userStockAccess databaseAccessUserManagement.UserStocksDataAccessInterface, stockDatabaseAccess databaseAccessStock.DatabaseAccessInterface, networkManager network.NetworkInterface) {
	_userStockAccess = userStockAccess
	_stockDatabaseAccess = stockDatabaseAccess
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

func getStockPortfolioHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	log.Println("[DEBUG] getStockPortfolioHandler invoked")

	userID := queryParams.Get("userID")
	if userID == "" {
		log.Println("[DEBUG] Error: missing userID in getStockPortfolioHandler")
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("[DEBUG] Extracted userID: %s", userID)

	stocks, err := _userStockAccess.GetUserStocks(userID)
	if err != nil {
		log.Printf("[DEBUG] Error retrieving stocks for userID %s: %v", userID, err)
		responseWriter.WriteHeader(http.StatusNotFound)
		return
	}
	log.Printf("[DEBUG] Retrieved %d stock records for userID %s", len(*stocks), userID)

	// Create custom response structure
	type StockPortfolioResponse struct {
		StockID       string    `json:"stock_id"`
		StockName     string    `json:"stock_name"`
		QuantityOwned int       `json:"quantity_owned"`
		UpdatedAt     time.Time `json:"updated_at"`
	}

	// Transform stocks into desired format
	portfolioResponse := make([]StockPortfolioResponse, 0)
	for _, stock := range *stocks {
		log.Printf("[DEBUG] Processing stock: ID=%s, Name=%s, Quantity=%d", stock.GetStockID(), stock.GetStockName(), stock.GetQuantity())
		if stock.GetQuantity() > 0 { // Only include stocks with quantity > 0
			portfolioResponse = append(portfolioResponse, StockPortfolioResponse{
				StockID:       stock.GetStockID(),
				StockName:     stock.GetStockName(),
				QuantityOwned: stock.GetQuantity(),
				UpdatedAt:     stock.GetUpdatedAt(),
			})
		}
	}
	log.Printf("[DEBUG] Transformed portfolio has %d entries", len(portfolioResponse))

	// Sort by stock name (if still needed)
	sort.SliceStable(portfolioResponse, func(i, j int) bool {
		return portfolioResponse[i].StockName > portfolioResponse[j].StockName
	})
	log.Println("[DEBUG] Sorted portfolio response by stock name")

	returnVal := network.ReturnJSON{
		Success: true,
		Data:    portfolioResponse,
	}

	stocksJSON, err := json.Marshal(returnVal)
	if err != nil {
		log.Printf("[DEBUG] Error marshalling JSON: %v", err)
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("[DEBUG] Marshalled JSON response: %s", string(stocksJSON))

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.Write(stocksJSON)
	log.Println("[DEBUG] Response written successfully")
}

func addStockToUser(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
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

	// Get the stock name using the map-based lookup
	stockName, err := getStockName(stockRequest.StockID)
	if err != nil {
		log.Printf("WARNING: Could not find stock name for ID %s: %v", stockRequest.StockID, err)
		stockName = "Unknown" // Fallback to "Unknown" if not found
	}

	newUserStock := userStock.New(userStock.NewUserStockParams{
		UserID:    userID,
		StockID:   stockRequest.StockID,
		Quantity:  stockRequest.Quantity,
		StockName: stockName, // Use the retrieved stock name
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

func getStockName(stockID string) (string, error) {
	log.Printf("[getStockName Debug] Looking up stock with ID: %s", stockID)
	stocks, err := _stockDatabaseAccess.GetAll()
	if err != nil {
		log.Printf("[getStockName Debug] Error getting stocks: %v", err)
		return "", err
	}
	stockIDToName := make(map[string]string)
	for _, stock := range *stocks {
		id := stock.GetId()
		name := stock.GetName()
		stockIDToName[id] = name
		log.Printf("[getStockName Debug] Stock found: ID=%s, Name=%s", id, name)
	}
	if name, exists := stockIDToName[stockID]; exists {
		log.Printf("[getStockName Debug] Found stock name for ID %s: %s", stockID, name)
		return name, nil
	}
	log.Printf("[getStockName Debug] Stock not found with ID: %s", stockID)
	return "", fmt.Errorf("stock not found with ID: %s", stockID)
}
