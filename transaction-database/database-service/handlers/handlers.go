package transactionDatabaseHandlers

import (
	"Shared/entities/transaction"
	"Shared/network"
	databaseServiceTransaction "databaseServiceTransaction/database-connection"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"time"

	"gorm.io/gorm"
)

var _databaseManager databaseServiceTransaction.DatabaseServiceInterface
var _networkManager network.NetworkInterface

func InitalizeHandlers(
	networkManager network.NetworkInterface, databaseManager databaseServiceTransaction.DatabaseServiceInterface) {
	_databaseManager = databaseManager
	_networkManager = networkManager

	//Add handlers
	networkManager.AddHandleFuncProtected(network.HandlerParams{Pattern: os.Getenv("transaction_route") + "/getStockTransactions", Handler: GetStockTransactions})
	networkManager.AddHandleFuncProtected(network.HandlerParams{Pattern: os.Getenv("transaction_route") + "/getWalletTransactions", Handler: getWalletTransactions})
	networkManager.AddHandleFuncUnprotected(network.HandlerParams{Pattern: "cancelStockTransaction/", Handler: cancelStockTransactionHandler})
	network.CreateNetworkEntityHandlers[*transaction.StockTransaction](_networkManager, os.Getenv("TRANSACTION_DATABASE_SERVICE_STOCK_ROUTE"), _databaseManager.StockTransactions(), transaction.ParseStockTransaction, transaction.ParseStockTransactionList)
	network.CreateNetworkEntityHandlers[*transaction.WalletTransaction](_networkManager, os.Getenv("TRANSACTION_DATABASE_SERVICE_WALLET_ROUTE"), _databaseManager.WalletTransactions(), transaction.ParseWalletTransaction, transaction.ParseWalletTransactionList)
	http.HandleFunc("/health", healthHandler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Simple check: you might expand this to test database connectivity, etc.
	w.WriteHeader(http.StatusOK)
	//fmt.Println(w, "OK")
}

func GetStockTransactions(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	transactions, err := _databaseManager.StockTransactions().GetAll()
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Create formatted response structure
	type FormattedStockTransaction struct {
		StockTxID       string    `json:"stock_tx_id"`
		ParentStockTxID *string   `json:"parent_stock_tx_id"` // Using pointer for null values
		StockID         string    `json:"stock_id"`
		WalletTxID      *string   `json:"wallet_tx_id"` // Using pointer for null values
		OrderStatus     string    `json:"order_status"`
		IsBuy           bool      `json:"is_buy"`
		OrderType       string    `json:"order_type"`
		StockPrice      float64   `json:"stock_price"`
		Quantity        int       `json:"quantity"`
		Timestamp       time.Time `json:"time_stamp"`
	}

	// Format transactions
	formattedTransactions := make([]FormattedStockTransaction, 0)
	for _, tx := range *transactions {
		tx.SetStockTXID() // Ensure ID is set

		// Create formatted transaction
		formatted := FormattedStockTransaction{
			StockTxID:   tx.GetId(),
			StockID:     tx.GetStockID(),
			OrderStatus: tx.GetOrderStatus(),
			IsBuy:       tx.GetIsBuy(),
			OrderType:   tx.GetOrderType(),
			StockPrice:  tx.GetStockPrice(),
			Quantity:    tx.GetQuantity(),
			Timestamp:   tx.GetTimestamp(),
		}

		// Handle nullable fields
		if parentID := tx.GetParentStockTransactionID(); parentID != "" {
			formatted.ParentStockTxID = &parentID
		}
		if walletID := tx.GetWalletTransactionID(); walletID != "" {
			formatted.WalletTxID = &walletID
		}

		formattedTransactions = append(formattedTransactions, formatted)
	}

	// Sort by timestamp
	sort.SliceStable(formattedTransactions, func(i, j int) bool {
		return formattedTransactions[i].Timestamp.Before(formattedTransactions[j].Timestamp)
	})

	// Create response
	returnVal := network.ReturnJSON{
		Success: true,
		Data:    formattedTransactions,
	}
	transactionsJSON, err := json.Marshal(returnVal)
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	responseWriter.Write(transactionsJSON)
}

// Expected input is a stock ID in the body of the request
// we're expecting {"StockID":"{id value}"}
func getWalletTransactions(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	transactions, err := _databaseManager.WalletTransactions().GetAll()
	if err != nil {
		fmt.Println("error: ", err.Error())
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	for _, transaction := range *transactions {
		//making sure the wallet_tx_id is set
		transaction.SetWalletTXID()
	}
	//sort transactions by timestamp. Oldest to newest
	sort.SliceStable((*transactions), func(i, j int) bool {
		return (*transactions)[i].GetTimestamp().Before((*transactions)[j].GetTimestamp())
	})
	returnVal := network.ReturnJSON{
		Success: true,
		Data:    transactions,
	}
	transactionsJSON, err := json.Marshal(returnVal)
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	responseWriter.Write(transactionsJSON)
}

func cancelStockTransactionHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	stockTransaction, err := _databaseManager.StockTransactions().GetByID(queryParams.Get("id"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		responseWriter.WriteHeader(http.StatusNotFound)
		return
	}
	stockTransaction.SetOrderStatus("CANCELLED")
	err = _databaseManager.StockTransactions().Update(stockTransaction.Updates)
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	responseWriter.WriteHeader(http.StatusOK)
}
