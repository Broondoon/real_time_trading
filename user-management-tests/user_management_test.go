package main

import (
	"Shared/network"
	"encoding/json"
	"fmt"
	"log"
	"testing"
)

func TestGetWalletBalance(t *testing.T) {
	network := network.NewNetwork()
	userID := "test-user-1"

	response, err := network.UserManagement().Get("getWalletBalance", map[string]string{"userID": userID})
	if err != nil {
		log.Fatalf("Failed to get wallet balance: %v", err)
	}
	var result map[string]float64
	if err := json.Unmarshal(response, &result); err != nil {
		log.Fatalf("Failed to parse wallet balance response: %v", err)
	}

	fmt.Printf("Wallet Balance for user %s: $%.2f\n", userID, result["balance"])
}
