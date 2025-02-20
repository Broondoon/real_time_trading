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
	network.CreateNetworkEntityHandlers[*transaction.StockTransaction](_networkManager, os.Getenv("TRANSACTION_DATABASE_SERVICE_STOCK_ROUTE"), _databaseManager.StockTransactions(), transaction.ParseStockTransaction)
	network.CreateNetworkEntityHandlers[*transaction.WalletTransaction](_networkManager, os.Getenv("TRANSACTION_DATABASE_SERVICE_WALLET_ROUTE"), _databaseManager.WalletTransactions(), transaction.ParseWalletTransaction)
	http.HandleFunc("/health", healthHandler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Simple check: you might expand this to test database connectivity, etc.
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func GetStockTransactions(responseWriter http.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	transactions, err := _databaseManager.StockTransactions().GetAll()
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	transactionsJSON, err := json.Marshal(transactions)
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	responseWriter.Write(transactionsJSON)
}

// Expected input is a stock ID in the body of the request
// we're expecting {"StockID":"{id value}"}
func getWalletTransactions(responseWriter http.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	transactions, err := _databaseManager.WalletTransactions().GetAll()
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	transactionsJSON, err := json.Marshal(transactions)
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	responseWriter.Write(transactionsJSON)
}

func cancelStockTransactionHandler(responseWriter http.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	stockTransaction, err := _databaseManager.StockTransactions().GetByID(queryParams.Get("id"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		responseWriter.WriteHeader(http.StatusNotFound)
		return
	}
	stockTransaction.SetOrderStatus("CANCELLED")
	err = _databaseManager.StockTransactions().Update(stockTransaction)
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	responseWriter.WriteHeader(http.StatusOK)
}
