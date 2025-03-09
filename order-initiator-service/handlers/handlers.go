package OrderInitiatorService

import (
	"Shared/entities/order"
	"Shared/entities/transaction"
	userStock "Shared/entities/user-stock"
	"Shared/network"
	subfunctions "Shared/subfunctions/Multithreading"
	"databaseAccessTransaction"
	"databaseAccessUserManagement"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"gorm.io/gorm"
)

const TIMEOUT = 2 * time.Second

var _databaseAccess databaseAccessTransaction.DatabaseAccessInterface
var _databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface
var _networkHttpManager network.NetworkInterface
var _networkQueueManager network.NetworkInterface

var _bulkRoutineStockOrderCheckUserStocks subfunctions.BulkRoutineInterface[*StockOrderBulk]
var _bulkRoutineStockOrderUpdateUserStocks subfunctions.BulkRoutineInterface[*StockOrderBulk]
var _bulkRoutineCreateStockOrderTransactions subfunctions.BulkRoutineInterface[*StockOrderBulk]

type StockOrderBulk struct {
	StockOrder     order.StockOrderInterface
	UserStock      userStock.UserStockInterface
	ResponseWriter network.ResponseWriter
	userId         string
	timeStamp      string
}

func InitalizeHandlers(
	networkHttpManager network.NetworkInterface, networkQueueManager network.NetworkInterface, databaseAccess databaseAccessTransaction.DatabaseAccessInterface, databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface) {
	_databaseAccess = databaseAccess
	_databaseAccessUser = databaseAccessUser
	_networkHttpManager = networkHttpManager
	_networkQueueManager = networkQueueManager

	_bulkRoutineStockOrderCheckUserStocks = subfunctions.NewBulkRoutine[*StockOrderBulk](&subfunctions.BulkRoutineParams[*StockOrderBulk]{
		Routine: checkUserStocks,
	})

	_bulkRoutineStockOrderUpdateUserStocks = subfunctions.NewBulkRoutine[*StockOrderBulk](&subfunctions.BulkRoutineParams[*StockOrderBulk]{
		Routine: updateUserStocks,
	})

	_bulkRoutineCreateStockOrderTransactions = subfunctions.NewBulkRoutine[*StockOrderBulk](&subfunctions.BulkRoutineParams[*StockOrderBulk]{
		Routine: placeStockOrderResponse,
	})

	//Add handlers
	_networkHttpManager.AddHandleFuncProtected(network.HandlerParams{Pattern: os.Getenv("engine_route") + "/placeStockOrder", Handler: placeStockOrderHandler})
	_networkHttpManager.AddHandleFuncProtected(network.HandlerParams{Pattern: os.Getenv("engine_route") + "/cancelStockTransaction", Handler: cancelStockTransactionHandler})
	http.HandleFunc("/health", healthHandler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Simple check: you might expand this to test database connectivity, etc.
	w.WriteHeader(http.StatusOK)
	//log.Println(w, "OK")
}

func placeStockOrderHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	log.Println("Placing stock order")
	stockOrder, err := order.Parse(data)
	if err != nil {
		log.Println("Error: ", err.Error())
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	stockOrder.SetUserID(queryParams.Get("userID"))
	stockOrderCarry := &StockOrderBulk{
		StockOrder:     stockOrder,
		ResponseWriter: responseWriter,
		userId:         queryParams.Get("userID"),
	}
	_bulkRoutineStockOrderCheckUserStocks.Insert(stockOrderCarry)
}

func checkUserStocks(data *[]*StockOrderBulk, TransferParams any) error {
	log.Println("Checking user stocks")
	// bul routine, taking in stock order.
	//then we organize the stock order's by USer IDS
	//then we run the bulk routine on user stocks. That will give us back
	//map stock orders by user id.
	ordersByUserId := make(map[string][]*StockOrderBulk)
	userIds := make([]string, 0)
	for _, stockOrder := range *data {
		if stockOrder.StockOrder.GetIsBuy() {
			_bulkRoutineCreateStockOrderTransactions.Insert(stockOrder)
		} else {
			ordersByUserId[stockOrder.userId] = append(ordersByUserId[stockOrder.userId], stockOrder)
			userIds = append(userIds, stockOrder.userId)
		}
	}

	handleSellOrders := func(userID string, sellerStockPortfolio *[]userStock.UserStockInterface, errorCode int) {
		if errorCode != 0 {
			for _, stockOrder := range ordersByUserId[userID] {
				if errorCode == http.StatusNotFound {
					log.Printf("user %s not found", userID)
					stockOrder.ResponseWriter.WriteHeader(http.StatusNotFound)
				} else {
					log.Printf("failed to get user stocks for user %s", userID)
					stockOrder.ResponseWriter.WriteHeader(http.StatusInternalServerError)
				}
			}
			return
		}
		sellOrders := ordersByUserId[userID]
		for _, stockOrder := range sellOrders {
			// Find the stock in the seller's portfolio
			var sellerStock userStock.UserStockInterface
			for _, stock := range *sellerStockPortfolio {
				if stock.GetStockID() == stockOrder.StockOrder.GetStockID() {
					sellerStock = stock
					break
				}
			}

			// Verify seller has the stock and sufficient quantity
			if sellerStock == nil {
				log.Printf("seller does not own stock %s", stockOrder.StockOrder.GetStockID())
				stockOrder.ResponseWriter.WriteHeader(http.StatusBadRequest)
				continue
			}
			if sellerStock.GetQuantity() < stockOrder.StockOrder.GetQuantity() {
				log.Printf("insufficient stock quantity: has %d, wants to sell %d\n",
					sellerStock.GetQuantity(), stockOrder.StockOrder.GetQuantity())
				stockOrder.ResponseWriter.WriteHeader(http.StatusBadRequest)
				continue
			}

			// Deduct the quantity from seller's portfolio but keep the record
			//need to bulkify this...
			sellerStock.UpdateQuantity(-stockOrder.StockOrder.GetQuantity())
			//what if we create a map of user to stock and subtract the quantity from the map, creatinga  subtraction value that we apply at the end.
			stockOrder.UserStock = sellerStock
			_bulkRoutineStockOrderUpdateUserStocks.Insert(stockOrder)
		}
	}
	err := _databaseAccessUser.UserStock().GetUserStocksBulk(userIds, handleSellOrders)
	if err != nil {
		for _, responseWriter := range *data {
			responseWriter.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		}
		log.Printf("failed to get user stocks: %v", err)
		return fmt.Errorf("failed to get user stocks: %v", err)
	}
	return nil
}

func updateUserStocks(data *[]*StockOrderBulk, TransferParams any) error {
	log.Println("Updating user stocks")
	//map user stocks by id and by stock id
	//then map then map them to the stock orders
	//then we
	userStocks := []userStock.UserStockInterface{}
	for _, stockOrder := range *data {
		userStocks = append(userStocks, stockOrder.UserStock)
	}
	//bulk update user stocks
	//TODO create a setup that errors out only specific parts of the update, not the entire thing.
	errorList, err := _databaseAccessUser.UserStock().UpdateBulk(&userStocks)
	if err != nil {
		log.Printf("Transaction Error failed to update user stocks: %v", err)
		for _, responseWriter := range *data {
			responseWriter.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		}
		return fmt.Errorf("failed to update user stocks: %v", err)
	}

	for _, stockOrder := range *data {
		if errorCode := errorList[stockOrder.UserStock.GetId()]; errorCode != 0 {
			log.Println("Stock order with ID: ", stockOrder.UserStock.GetId(), " has Error code: ", errorCode)
			if errorCode == http.StatusNotFound {
				log.Printf("user stock %s not found", stockOrder.UserStock.GetId())
				stockOrder.ResponseWriter.WriteHeader(http.StatusNotFound)
			} else {
				log.Printf("failed to update user stock %s", stockOrder.UserStock.GetId())
				stockOrder.ResponseWriter.WriteHeader(http.StatusInternalServerError)
			}
			continue
		}
		_bulkRoutineCreateStockOrderTransactions.Insert(stockOrder)
	}
	return nil
}

func placeStockOrderResponse(data *[]*StockOrderBulk, TransferParams any) error {
	log.Println("Creating stock order transactions")
	bulkTransactions := make([]transaction.StockTransactionInterface, len(*data))
	for _, stockOrder := range *data {
		newTransaction := transaction.NewStockTransaction(transaction.NewStockTransactionParams{
			StockOrder:  stockOrder.StockOrder,
			OrderStatus: "IN_PROGRESS",
			TimeStamp:   time.Now(),
		})
		bulkTransactions = append(bulkTransactions, newTransaction)
		//get string version of time stamp

		stockOrder.timeStamp = newTransaction.GetTimestamp().String()
	}
	createdTransactions, errList, err := _databaseAccess.StockTransaction().CreateBulk(&bulkTransactions)
	if err != nil {
		for _, responseWriter := range *data {
			responseWriter.ResponseWriter.WriteHeader(http.StatusInternalServerError)
			log.Printf("failed to create transactions: %v", err)
			return fmt.Errorf("failed to create transactions: %v", err)
		}
	}
	stockOrdersByTimeStamp := make(map[string][]*StockOrderBulk)
	for _, stockOrder := range *data {
		stockOrdersByTimeStamp[stockOrder.timeStamp] = append(stockOrdersByTimeStamp[stockOrder.timeStamp], stockOrder)
	}
	for timeStamp, err := range errList {
		if err != 0 {
			for _, stockOrder := range stockOrdersByTimeStamp[timeStamp] {
				log.Printf("failed to create transaction %s", timeStamp)
				stockOrder.ResponseWriter.WriteHeader(http.StatusInternalServerError)
			}
		}
	}

	for _, createdTransaction := range *createdTransactions {
		reconstructedStockOrder := order.New(order.NewStockOrderParams{
			StockID:   createdTransaction.GetStockID(),
			IsBuy:     createdTransaction.GetIsBuy(),
			OrderType: createdTransaction.GetOrderType(),
			Quantity:  createdTransaction.GetQuantity(),
			Price:     createdTransaction.GetStockPrice(),
			UserID:    createdTransaction.GetUserID(),
		})
		_, err = _networkQueueManager.MatchingEngine().Post("placeStockOrder", reconstructedStockOrder)
		if err != nil {
			for _, responseWriter := range *data {
				responseWriter.ResponseWriter.WriteHeader(http.StatusInternalServerError)
				log.Printf("failed to send to matching engine: %v", err)
				return fmt.Errorf("failed to send to matching engine: %v", err)
			}
		}
	}
	for _, responseWriter := range *data {
		returnVal := network.ReturnJSON{
			Success: true,
			Data:    nil,
		}
		returnValJSON, err := json.Marshal(returnVal)
		if err != nil {
			log.Printf("failed to marshal return value: %v", err)
			responseWriter.ResponseWriter.WriteHeader(http.StatusInternalServerError)
			continue
		}
		responseWriter.ResponseWriter.Write(returnValJSON)
	}
	return nil
}

func cancelStockTransactionHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	log.Println("Cancelling stock transaction")
	var stockID network.StockTransactionID
	err := json.Unmarshal(data, &stockID)
	if err != nil {
		log.Println("Error: ", err.Error())
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	err = cancelStockTransaction(stockID.StockTransactionID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		responseWriter.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		log.Println("Error: ", err.Error())
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
	err := _networkHttpManager.Transactions().Patch("cancelStockTransaction", id)
	if err != nil {
		log.Println("Error: ", err.Error())
		return err
	}

	_, err = _networkQueueManager.MatchingEngine().Delete("deleteOrder/" + id)
	if err != nil {
		log.Println("Error: ", err.Error())
		return err
	}
	return nil

}
