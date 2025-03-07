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
	network.CreateNetworkEntityHandlers[*transaction.StockTransaction](_networkManager, os.Getenv("TRANSACTION_DATABASE_SERVICE_STOCK_ROUTE"), _databaseManager.StockTransactions(), transaction.ParseStockTransaction)
	network.CreateNetworkEntityHandlers[*transaction.WalletTransaction](_networkManager, os.Getenv("TRANSACTION_DATABASE_SERVICE_WALLET_ROUTE"), _databaseManager.WalletTransactions(), transaction.ParseWalletTransaction)
	http.HandleFunc("/health", healthHandler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	//fmt.Println(w, "OK")
}

func GetStockTransactions(responseWriter http.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	transactions, err := _databaseManager.StockTransactions().GetByForeignID("user_id", queryParams.Get("userID"))


	if err != nil {
        println("Had an error. Error: ", err.Error())
        responseWriter.WriteHeader(http.StatusInternalServerError)
        return
    }
    for _, transaction := range *transactions {
        json, err := transaction.ToJSON()
        if err != nil {
            println("Had an error. Error: ", err.Error())
            responseWriter.WriteHeader(http.StatusInternalServerError)
            return
        }
        fmt.Println(string(json))
    }

	
	// Formatted response structure to match the expected output
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
// ** ^ ?? Need to clarify this with group - The Expected Header input should be {"token":<user1Token>} or {"token":<compToken>}
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
/* Example of Expected Output:
"success":true,
"data":[{"wallet_tx_id"":<googleW
alletTxId>,
"stock_tx_id":<googleStockTxId>,
"is_debit":true, "amount":1350,
"time_stamp":<timestamp>}]
*/
func getWalletTransactions(responseWriter http.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
    walletTransactions, err := _databaseManager.WalletTransactions().GetByForeignID("user_id", queryParams.Get("userID"))

    if err != nil {
        println("Had an error. Error: ", err.Error())
        responseWriter.WriteHeader(http.StatusInternalServerError)
        return
    }
    for _, transaction := range *walletTransactions {
        json, err := transaction.ToJSON()
        if err != nil {
            println("Had an error. Error: ", err.Error())
            responseWriter.WriteHeader(http.StatusInternalServerError)
            return
        }
        fmt.Println(string(json))
    }

    // Formatted response structure to match the expected output
    type FormattedWalletTransaction struct {
        WalletTxID string   `json:"wallet_tx_id"`
        StockTxID  string    `json:"stock_tx_id"`
        IsDebit    bool      `json:"is_debit"`
        Amount     float64   `json:"amount"`
        Timestamp  time.Time `json:"time_stamp"`
    }

    // Format transactions
    formattedTransactions := make([]FormattedWalletTransaction, 0)
    for _, tx := range *walletTransactions {
        tx.SetWalletTXID() // ensure the wallet_tx_id is set
        

        // Create formatted transaction
        formatted := FormattedWalletTransaction{
            WalletTxID: tx.GetId(),  // Get the ID from the wallet transaction
            StockTxID:  tx.GetStockTransactionID(),
            IsDebit:    tx.GetIsDebit(),
            Amount:     tx.GetAmount(),
            Timestamp:  tx.GetTimestamp(),
        }

        formattedTransactions = append(formattedTransactions, formatted)
    }

    //sort transactions by timestamp. Oldest to newest
    sort.SliceStable((formattedTransactions), func(i, j int) bool {
        return formattedTransactions[i].Timestamp.Before(formattedTransactions[j].Timestamp)
    })

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
