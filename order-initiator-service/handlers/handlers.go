package OrderInitiatorService

import (
	"Shared/entities/order"
	"Shared/entities/transaction"
	userStock "Shared/entities/user-stock"
	"Shared/network"
	"databaseAccessTransaction"
	"databaseAccessUserManagement"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"gorm.io/gorm"
)

var _databaseAccess databaseAccessTransaction.DatabaseAccessInterface
var _databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface
var _networkManager network.NetworkInterface

func InitalizeHandlers(
	networkManager network.NetworkInterface, databaseAccess databaseAccessTransaction.DatabaseAccessInterface, databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface) {
	_databaseAccess = databaseAccess
	_databaseAccessUser = databaseAccessUser
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

	if !stockOrder.GetIsBuy() {
		// Get seller's current stock holdings
		sellerStockPortfolio, err := _databaseAccessUser.UserStock().GetUserStocks(stockOrder.GetUserID())
		if err != nil {
			return fmt.Errorf("failed to get seller stocks: %v", err)
		}

		// Find the stock in the seller's portfolio
		var sellerStock userStock.UserStockInterface
		for _, stock := range *sellerStockPortfolio {
			if stock.GetStockID() == stockOrder.GetStockID() {
				sellerStock = stock
				break
			}
		}

		// Verify seller has the stock and sufficient quantity
		if sellerStock == nil {
			return fmt.Errorf("seller does not own stock %s", stockOrder.GetStockID())
		}
		if sellerStock.GetQuantity() < stockOrder.GetQuantity() {
			return fmt.Errorf("insufficient stock quantity: has %d, wants to sell %d",
				sellerStock.GetQuantity(), stockOrder.GetQuantity())
		}

		// Deduct the quantity from seller's portfolio but keep the record
		newQuantity := sellerStock.GetQuantity() - stockOrder.GetQuantity()
		sellerStock.SetQuantity(newQuantity)
		err = _databaseAccessUser.UserStock().Update(sellerStock)
		if err != nil {
			return fmt.Errorf("failed to update seller stock quantity: %v", err)
		}
	}

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

/*
func cancelStockTransaction(id string) error {
    // Get the transaction details first
    transaction, err := _databaseAccess.StockTransaction().Get(id)
    if err != nil {
        return fmt.Errorf("failed to get transaction: %v", err)
    }

    // If it's a sell order, restore the seller's stock quantity
    if !transaction.GetIsBuy() {
        stockOrder := transaction.GetStockOrder()
        sellerID := stockOrder.GetUserID()
        stockID := stockOrder.GetStockID()
        quantity := stockOrder.GetQuantity()

        // Get seller's current stock holdings
        sellerStock, err := _databaseAccessUser.UserStock().GetUserStock(sellerID, stockID)
        if err != nil {
            return fmt.Errorf("failed to get seller stock: %v", err)
        }

        // Restore the quantity
        sellerStock.SetQuantity(sellerStock.GetQuantity() + quantity)
        err = _databaseAccessUser.UserStock().Update(sellerStock)
        if err != nil {
            return fmt.Errorf("failed to restore seller stock quantity: %v", err)
        }
    }

    // Continue with existing cancellation logic
    _, err = _networkManager.Transactions().Put("cancelStockTransaction/"+id, nil)
    if err != nil {
        return err
    }

    _, err = _networkManager.MatchingEngine().Delete("deleteOrder/" + id)
    return err
}
*/
