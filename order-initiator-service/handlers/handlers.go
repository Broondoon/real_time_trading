package OrderInitiatorService

import (
	"Shared/entities/order"
	"Shared/entities/transaction"
	"Shared/network"
	"databaseAccessTransaction"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"time"

	"gorm.io/gorm"
)

var _databaseAccess databaseAccessTransaction.DatabaseAccessInterface
var _networkManager network.NetworkInterface

func InitalizeHandlers(
	networkManager network.NetworkInterface, databaseAccess databaseAccessTransaction.DatabaseAccessInterface) {
	_databaseAccess = databaseAccess
	_networkManager = networkManager

	//listen for placeStockOrder. Create a new stock Transaction, updatet he stock order id, pass it to the matching engine.
	//listen for cancelStockTransaction.

	//Add handlers
	networkManager.AddHandleFuncProtected(network.HandlerParams{Pattern: os.Getenv("engine_route") + "/placeStockOrder", Handler: placeStockOrderHandler})
	networkManager.AddHandleFuncProtected(network.HandlerParams{Pattern: os.Getenv("engine_route") + "/cancelStockTransaction", Handler: cancelStockTransactionHandler})
	http.HandleFunc("/health", healthHandler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Simple check: you might expand this to test database connectivity, etc.
	w.WriteHeader(http.StatusOK)
	//fmt.Println(w, "OK")
}

func placeStockOrderHandler(responseWriter http.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	println("Placing stock order")
	stockOrder, err := order.Parse(data)
	if err != nil {
		println("Error: ", err.Error())
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	stockOrder.SetUserID(queryParams.Get("userID"))
	err = placeStockOrder(stockOrder)
	if err != nil {
		println("Error: ", err.Error())
		responseWriter.WriteHeader(http.StatusInternalServerError)
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
}

func placeStockOrder(stockOrder order.StockOrderInterface) error {
	var err error
	transaction := transaction.NewStockTransaction(transaction.NewStockTransactionParams{
		StockOrder:  stockOrder,
		OrderStatus: "IN_PROGRESS",
		TimeStamp:   time.Now(),
	})

	createdTransaction, err := _databaseAccess.StockTransaction().Create(transaction)
	if err != nil {
		println("Error: ", err.Error())
		return err
	}
	stockOrder.SetId(createdTransaction.GetId())
	//pass to matching engine
	_, err = _networkManager.MatchingEngine().Post("placeStockOrder", stockOrder)
	return err
}

func cancelStockTransactionHandler(responseWriter http.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	println("Cancelling stock transaction")
	var stockID network.StockTransactionID
	err := json.Unmarshal(data, &stockID)
	if err != nil {
		println("Error: ", err.Error())
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	err = cancelStockTransaction(stockID.StockTransactionID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		responseWriter.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		println("Error: ", err.Error())
		responseWriter.WriteHeader(http.StatusInternalServerError)
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
}

func cancelStockTransaction(id string) error {
	//pass to matching engine
	_, err := _networkManager.Transactions().Put("cancelStockTransaction/"+id, nil)
	if err != nil {
		println("Error: ", err.Error())
		return err
	}

	_, err = _networkManager.MatchingEngine().Delete("deleteOrder/" + id)
	if err != nil {
		println("Error: ", err.Error())
		return err
	}
	return nil

}
