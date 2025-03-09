package main

import (
	networkHttp "Shared/network/http"
	"encoding/json"
	"log"
	"testing"
)

var userID = "6fd2fc6b-9142-4777-8b30-575ff6fa2460"
var stockId = "69e81793-1cc7-476f-a8ba-714fafcb3e5c"
var client = networkHttp.NewNetworkHttp().UserManagement()

func TestGetWalletBalance(t *testing.T) {
	queryParams := map[string]string{"userID": userID}
	response, err := client.Get("transaction/getWalletBalance", queryParams)

	if err != nil {
		t.Fatalf("Failed to get wallet balance: %v", err)
	}

	t.Logf("Response: %s", string(response))

	var balanceResponse struct {
		Balance float64 `json:"balance"`
	}
	if err := json.Unmarshal(response, &balanceResponse); err != nil {
		t.Fatalf("Failed to parse wallet balance response: %v", err)
	}

	if balanceResponse.Balance <= 0 {
		t.Errorf("Expected positive balance, got: %.2f", balanceResponse.Balance)
	}
}

func TestAddMoneyToWallet(t *testing.T) {
	payload := map[string]interface{}{
		"amount": 100,
	}

	response, err := client.Post("transaction/addMoneyToWallet", payload)
	if err != nil {
		t.Fatalf("Failed to add money to wallet: %v", err)
	}
	log.Println(string(response))

	var result struct {
		Success bool        `json:"success"`
		Data    interface{} `json:"data"`
	}
	if err := json.Unmarshal(response, &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success to be true, got false")
	}
	if result.Data != nil {
		t.Errorf("Expected data to be null, got: %v", result.Data)
	}
}

func TestGetStockPortfolio(t *testing.T) {
	queryParams := map[string]string{"userID": userID}
	response, err := client.Get("transaction/getStockPortfolio", queryParams)
	if err != nil {
		t.Fatalf("Failed to get stock portfolio: %v", err)
	}

	t.Logf("Response: %s", string(response))

	var portfolioResponse struct {
		Success bool `json:"success"`
		Data    []struct {
			StockID   string `json:"stock_id"`
			StockName string `json:"stock_name"`
			Quantity  int    `json:"quantity_owned"`
		} `json:"data"`
	}

	if err := json.Unmarshal(response, &portfolioResponse); err != nil {
		t.Fatalf("Failed to parse stock portfolio response: %v", err)
	}

	if !portfolioResponse.Success {
		t.Errorf("Expected success to be true, got false")
	}

	if len(portfolioResponse.Data) == 0 {
		t.Errorf("Expected at least one stock in portfolio, got none.")
	}
}

func TestAddStockToUser(t *testing.T) {
	requestBody := map[string]interface{}{
		"stock_id": stockId,
		"quantity": 10,
	}

	response, err := client.Post("setup/addStockToUser", requestBody)
	if err != nil {
		t.Fatalf("Failed to add stock to user: %v", err)
	}

	t.Logf("Response: %s", string(response))

	var addStockResponse struct {
		Success bool        `json:"success"`
		Data    interface{} `json:"data"`
	}

	if err := json.Unmarshal(response, &addStockResponse); err != nil {
		t.Fatalf("Failed to parse add stock response: %v", err)
	}

	if !addStockResponse.Success {
		t.Errorf("Expected success to be true, got false")
	}

	if addStockResponse.Data != nil {
		t.Errorf("Expected data to be null, got: %v", addStockResponse.Data)
	}
}
