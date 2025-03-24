package transactionDatabaseHandlers

import (
	"Shared/entities/transaction"
	"Shared/network"
	subfunctions "Shared/subfunctions/Multithreading"
	databaseServiceTransaction "databaseServiceTransaction/database-connection"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"time"

	"gorm.io/gorm"
)

var _databaseManager databaseServiceTransaction.DatabaseServiceInterface
var _networkManager network.NetworkInterface

var _bulkGetStockTransactions subfunctions.BulkRoutineInterface[*TransactionBulk]
var _bulkGetWalletTransactions subfunctions.BulkRoutineInterface[*TransactionBulk]

type TransactionBulk struct {
	userId string
	network.ResponseWriter
}

func InitalizeHandlers(
	networkManager network.NetworkInterface, databaseManager databaseServiceTransaction.DatabaseServiceInterface) {
	_databaseManager = databaseManager
	_networkManager = networkManager

	_bulkGetStockTransactions = subfunctions.NewBulkRoutine(&subfunctions.BulkRoutineParams[*TransactionBulk]{
		Routine: getStockTransactionsBulk,
	})
	_bulkGetWalletTransactions = subfunctions.NewBulkRoutine(&subfunctions.BulkRoutineParams[*TransactionBulk]{
		Routine: getWalletTransactionsBulk,
	})

	//Add handlers
	networkManager.AddHandleFuncProtected(network.HandlerParams{Pattern: os.Getenv("transaction_route") + "/getStockTransactions", Handler: GetStockTransactions})
	networkManager.AddHandleFuncProtected(network.HandlerParams{Pattern: os.Getenv("transaction_route") + "/getWalletTransactions", Handler: getWalletTransactions})
	networkManager.AddHandleFuncUnprotected(network.HandlerParams{Pattern: "cancelStockTransaction/", Handler: cancelStockTransactionHandler})
	network.CreateNetworkEntityHandlers[*transaction.StockTransaction](_networkManager, os.Getenv("TRANSACTION_DATABASE_SERVICE_STOCK_ROUTE"), _databaseManager.StockTransactions(), transaction.ParseStockTransaction, transaction.ParseStockTransactionList)
	network.CreateNetworkEntityHandlers[*transaction.WalletTransaction](_networkManager, os.Getenv("TRANSACTION_DATABASE_SERVICE_WALLET_ROUTE"), _databaseManager.WalletTransactions(), transaction.ParseWalletTransaction, transaction.ParseWalletTransactionList)
	http.HandleFunc("/health", healthHandler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	//log.Println(w, "OK")
}

func GetStockTransactions(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	_bulkGetStockTransactions.Insert(&TransactionBulk{
		userId:         queryParams.Get("userID"),
		ResponseWriter: responseWriter,
	})
	// log.Println("Getting stock transactions")
	// log.Println("Data: ", string(data))
	// log.Println("Query Params: ", queryParams.Encode())
	// log.Println("Request Type: ", requestType)

}

func getStockTransactionsBulk(data *[]*TransactionBulk, TransferParams any) error {
	userIds := make([]string, len(*data))
	mapUserIds := make(map[string]network.ResponseWriter)
	for i, d := range *data {
		userIds[i] = d.userId
		mapUserIds[d.userId] = d.ResponseWriter
	}
	transactionList, err := _databaseManager.StockTransactions().GetByForeignIDBulk("UserID", userIds)
	transactionMap := make(map[string][]transaction.StockTransactionInterface)
	for _, transaction := range *transactionList {
		transactionMap[transaction.GetUserIDString()] = append(transactionMap[transaction.GetUserIDString()], transaction)
	}
	for userId, responseWriter := range mapUserIds {
		if errCode, ok := err[userId]; ok {
			if errors.Is(errCode, gorm.ErrRecordNotFound) {
				responseWriter.WriteHeader(http.StatusNotFound)
				continue
			} else {
				responseWriter.WriteHeader(http.StatusInternalServerError)
				continue
			}
		}
		if transactions, ok := transactionMap[userId]; ok {

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
			formattedTransactions := make([]FormattedStockTransaction, len(transactions))
			for i, tx := range transactions {
				tx.SetStockTXID() // Ensure ID is set
				// Create formatted transaction
				formatted := FormattedStockTransaction{
					StockTxID:   tx.GetIdString(),
					StockID:     tx.GetStockIDString(),
					OrderStatus: tx.GetOrderStatus(),
					IsBuy:       tx.GetIsBuy(),
					OrderType:   tx.GetOrderType(),
					StockPrice:  tx.GetStockPrice(),
					Quantity:    tx.GetQuantity(),
					Timestamp:   tx.GetTimestamp(),
				}

				// Handle nullable fields
				if parentID := tx.GetParentStockTransactionIDString(); parentID != "" {
					formatted.ParentStockTxID = &parentID
				}
				if walletTXID := tx.GetWalletTransactionIDString(); walletTXID != "" {
					formatted.WalletTxID = &walletTXID
				}

				formattedTransactions[i] = formatted
			}

			// Sort by timestamp
			sort.SliceStable(formattedTransactions, func(i, j int) bool {
				//looks like this wants to get Market orders before limit orders
				if formattedTransactions[i].OrderType >= formattedTransactions[j].OrderType {
					//looks like this wants to get Cancelled before Completed before In Progress
					if formattedTransactions[i].OrderStatus <= formattedTransactions[j].OrderStatus {
						return formattedTransactions[i].Timestamp.Before(formattedTransactions[j].Timestamp)
					}
				}
				return true
			})

			// Create response
			returnVal := network.ReturnJSON{
				Success: true,
				Data:    formattedTransactions,
			}
			transactionsJSON, err := json.Marshal(returnVal)
			if err != nil {
				responseWriter.WriteHeader(http.StatusInternalServerError)
				continue
			}
			log.Println("Transactions: ", string(transactionsJSON))
			responseWriter.Write(transactionsJSON)
		} else {
			responseWriter.WriteHeader(http.StatusNotFound)
		}
	}
	return nil
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
func getWalletTransactions(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	_bulkGetWalletTransactions.Insert(&TransactionBulk{
		userId:         queryParams.Get("userID"),
		ResponseWriter: responseWriter,
	})
	// log.Println("Getting wallet transactions")
	// log.Println("Data: ", string(data))
	// log.Println("Query Params: ", queryParams.Encode())
	// log.Println("Request Type: ", requestType)

}

func getWalletTransactionsBulk(data *[]*TransactionBulk, TransferParams any) error {
	userIds := make([]string, len(*data))
	mapUserIds := make(map[string]network.ResponseWriter)
	for i, d := range *data {
		userIds[i] = d.userId
		mapUserIds[d.userId] = d.ResponseWriter
	}
	transactionList, err := _databaseManager.WalletTransactions().GetByForeignIDBulk("UserID", userIds)
	transactionMap := make(map[string][]transaction.WalletTransactionInterface)
	for _, transaction := range *transactionList {
		transactionMap[transaction.GetUserIDString()] = append(transactionMap[transaction.GetUserIDString()], transaction)
	}
	for userId, responseWriter := range mapUserIds {
		if errCode, ok := err[userId]; ok {
			if errors.Is(errCode, gorm.ErrRecordNotFound) {
				responseWriter.WriteHeader(http.StatusNotFound)
				continue
			} else {
				responseWriter.WriteHeader(http.StatusInternalServerError)
				continue
			}
		}
		if walletTransactions, ok := transactionMap[userId]; ok {

			// Formatted response structure to match the expected output
			type FormattedWalletTransaction struct {
				WalletTxID string    `json:"wallet_tx_id"`
				StockTxID  string    `json:"stock_tx_id"`
				IsDebit    bool      `json:"is_debit"`
				Amount     float64   `json:"amount"`
				Timestamp  time.Time `json:"time_stamp"`
			}

			// Format transactions
			formattedTransactions := make([]FormattedWalletTransaction, len(walletTransactions))
			for i, tx := range walletTransactions {
				tx.SetWalletTXID() // ensure the wallet_tx_id is set

				// Create formatted transaction
				formatted := FormattedWalletTransaction{
					WalletTxID: tx.GetIdString(), // Get the ID from the wallet transaction
					StockTxID:  tx.GetStockTransactionIDString(),
					IsDebit:    tx.GetIsDebit(),
					Amount:     tx.GetAmount(),
					Timestamp:  tx.GetTimestamp(),
				}

				formattedTransactions[i] = formatted
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
				continue
			}
			responseWriter.Write(transactionsJSON)
		} else {
			responseWriter.WriteHeader(http.StatusNotFound)
		}
	}
	return nil
}

func cancelStockTransactionHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	stockTransaction, err := _databaseManager.StockTransactions().GetByID(queryParams.Get("id"))

	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Println("Stock transaction not found")
		responseWriter.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		log.Println("Failed to get stock transaction, error: ", err.Error())
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	stockTransaction.SetOrderStatus("CANCELLED")
	errList := _databaseManager.StockTransactions().Update(*stockTransaction.Updates)
	if err := errList["transaction"]; err != nil {
		log.Println("Failed to update stock transaction, error: ", err.Error())
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		for _, err := range errList {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Println("Stock transaction not found")
				responseWriter.WriteHeader(http.StatusNotFound)
				return
			} else {
				log.Println("Failed to update stock transaction, error: ", err.Error())
				responseWriter.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
	responseWriter.WriteHeader(http.StatusOK)
}
