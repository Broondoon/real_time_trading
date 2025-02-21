package main

import (
	"Shared/network"
	"bytes"
	"encoding/json"
	"log"
	"testing"
)

var userID = "test-user-1"
var client = network.NewNetwork().UserManagement()
var stockClient = network.NewNetwork().UserManagementDatabase()

func TestAddMoneyToWallet(t *testing.T) {
	userID := "12345657"
	payload := map[string]interface{}{
		"amount": 100.00,
	}

	requestBody, _ := json.Marshal(payload)

	response, err := client.Post("transaction/addMoneyToWallet?userID="+userID, bytes.NewReader(requestBody))
	if err != nil {
		t.Fatalf("Failed to add money to wallet: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	log.Printf("Response: %v\n", result)

	if result["message"] != "Money added successfully" {
		t.Errorf("Expected success message, got: %v", result["message"])
	}
}

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

func TestGetStockPortfolio(t *testing.T) {
	queryParams := map[string]string{"userID": userID}
	response, err := stockClient.Get("transaction/getStockPortfolio", queryParams)

	if err != nil {
		t.Fatalf("Failed to get stock portfolio: %v", err)
	}

	t.Logf("Response: %s", string(response))

	var stockResponse []struct {
		StockID  string `json:"stock_id"`
		Quantity int    `json:"quantity"`
	}
	if err := json.Unmarshal(response, &stockResponse); err != nil {
		t.Fatalf("Failed to parse stock portfolio response: %v", err)
	}

	if len(stockResponse) == 0 {
		t.Errorf("Expected at least one stock in portfolio, got none.")
	}
}

func TestAddStockToUser(t *testing.T) {
	requestBody := map[string]interface{}{
		"stockID":  "AAPL",
		"quantity": 10,
	}

	requestJSON, _ := json.Marshal(requestBody)

	response, err := stockClient.Post("setup/addStockToUser?userID="+userID, bytes.NewReader(requestJSON))
	if err != nil {
		t.Fatalf("Failed to add stock to user: %v", err)
	}

	t.Logf("Response: %s", string(response))

	var addStockResponse struct {
		StockID  string `json:"stock_id"`
		Quantity int    `json:"quantity"`
	}
	if err := json.Unmarshal(response, &addStockResponse); err != nil {
		t.Fatalf("Failed to parse add stock response: %v", err)
	}

	if addStockResponse.StockID != "AAPL" || addStockResponse.Quantity != 10 {
		t.Errorf("Expected stock AAPL with quantity 10, got %s with quantity %d", addStockResponse.StockID, addStockResponse.Quantity)
	}
}
