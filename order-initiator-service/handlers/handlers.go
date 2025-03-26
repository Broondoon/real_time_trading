package OrderInitiatorService

import (
	"Shared/entities/order"
	"Shared/entities/transaction"
	userStock "Shared/entities/user-stock"
	"Shared/network"
	"Shared/objects"
	subfunctions "Shared/subfunctions/Multithreading"
	"databaseAccessStock"
	"databaseAccessTransaction"
	"databaseAccessUserManagement"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

const TIMEOUT = 2 * time.Second

var _databaseAccess databaseAccessTransaction.DatabaseAccessInterface
var _databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface
var _databaseAccessStock databaseAccessStock.DatabaseAccessInterface
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
}

func InitalizeHandlers(
	networkHttpManager network.NetworkInterface, networkQueueManager network.NetworkInterface, databaseAccess databaseAccessTransaction.DatabaseAccessInterface, databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface, databaseAccessStock databaseAccessStock.DatabaseAccessInterface) {
	_databaseAccess = databaseAccess
	_databaseAccessUser = databaseAccessUser
	_databaseAccessStock = databaseAccessStock
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
	// _networkHttpManager.AddHandleFuncProtected(network.HandlerParams{Pattern: os.Getenv("setup_route") + "/createStock", Handler: AddNewStockHandler})
	_networkHttpManager.AddHandleFuncProtected(network.HandlerParams{Pattern: os.Getenv("engine_route") + "/placeStockOrder", Handler: placeStockOrderHandler})
	_networkHttpManager.AddHandleFuncProtected(network.HandlerParams{Pattern: os.Getenv("engine_route") + "/cancelStockTransaction", Handler: cancelStockTransactionHandler})
	http.HandleFunc("/health", healthHandler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Simple check: you might expand this to test database connectivity, etc.
	w.WriteHeader(http.StatusOK)
	//log.Println(w, "OK")
}

// func AddNewStockHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
// 	newStock, err := stock.Parse(data)
// 	//log.Println("Parsed Stock: ", newStock.GetId())
// 	if err != nil {
// 		log.Println("Error: ", err.Error())
// 		responseWriter.WriteHeader(http.StatusBadRequest)
// 		return
// 	}
// 	returnedStock, err := _databaseAccessStock.AddNewStock(newStock)
// 	if err != nil {
// 		log.Println("Error: ", err.Error())
// 		responseWriter.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}
// 	_bulkRoutineStockOrderCheckUserStocks[returnedStock.GetIdString()] = subfunctions.NewBulkRoutine[*StockOrderBulk](&subfunctions.BulkRoutineParams[*StockOrderBulk]{
// 		Routine:        checkUserStocks,
// 		TransferParams: returnedStock.GetIdString(),
// 	})

// 	returnVal := network.ReturnJSON{
// 		Success: true,
// 		Data:    network.StockID{StockID: newStock.GetIdString()},
// 	}
// 	returnValJSON, err := json.Marshal(returnVal)
// 	if err != nil {
// 		responseWriter.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}
// 	responseWriter.Write(returnValJSON)
// }

func placeStockOrderHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	//log.Println("Placing stock order")
	stockOrder, err := order.Parse(data)
	if err != nil {
		log.Println("Handler Error: ", err.Error())
		log.Println("Handler Data: ", string(data))
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	userIDString := queryParams.Get("userID")
	if userIDString == "" {
		log.Println("Handler Error: userID not provided")
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Println("Handler userID: ", userIDString)
	userUuid, err := uuid.Parse(strings.TrimSpace(userIDString))
	if err != nil {
		log.Println("Handler Error: ", err.Error(), " ID attempted to parse: ", strings.TrimSpace(userIDString))
		for key, value := range queryParams {
			for _, v := range value {
				log.Println("Handler Query Param: ", key, v)
			}
		}

		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	stockOrder.SetUserID(&userUuid)
	stockOrderCarry := &StockOrderBulk{
		StockOrder:     stockOrder,
		ResponseWriter: responseWriter,
		userId:         userIDString,
	}
	_bulkRoutineStockOrderCheckUserStocks.Insert(stockOrderCarry)
}

func checkUserStocks(data *[]*StockOrderBulk, TransferParams any) error {
	pairs := make([]objects.Pair, len(*data))
	ordersByPairs := make(map[string][]*StockOrderBulk)
	i := 0
	for _, stockOrderCarry := range *data {
		if stockOrderCarry.StockOrder.GetIsBuy() {
			_bulkRoutineCreateStockOrderTransactions.Insert(stockOrderCarry)
		} else {
			paired := objects.Pair{
				ID1: stockOrderCarry.StockOrder.GetUserIDString(),
				ID2: stockOrderCarry.StockOrder.GetStockIDString(),
			}
			if _, ok := ordersByPairs[paired.String()]; !ok {
				ordersByPairs[paired.String()] = []*StockOrderBulk{}
				pairs[i] = paired
				i++
			}
			ordersByPairs[paired.String()] = append(ordersByPairs[paired.String()], stockOrderCarry)
		}
	}

	if i == 0 {
		log.Println("DEBUG: GetUserStocksBulk called with empty userIDs")
		return nil
	}
	var (
		userStocks *[]userStock.UserStockInterface
		errList    = make(map[string]int)
		err        error
	)
	userStocks, errList, err = _databaseAccessUser.UserStock().GetByPairedIDBulk("UserID", "StockID", &pairs)
	//lets make a variant which is get by foregin ids. Get back multiple, then perform a function for each userId
	if err != nil {
		log.Println("Error fetching user stocks by foreign ID for userIDs: ", err)
		return err
	}

	userStocksByPairs := make(map[string]userStock.UserStockInterface)
	for _, userStock := range *userStocks {
		paired := objects.Pair{
			ID1: userStock.GetUserIDString(),
			ID2: userStock.GetStockIDString(),
		}
		userStocksByPairs[paired.String()] = userStock
	}
	for key, stockOrderCarries := range ordersByPairs {
		if err, ok := errList[key]; ok && err != 0 {
			for _, stockOrderCarry := range stockOrderCarries {
				if err == http.StatusNotFound {
					log.Printf("user %s not found", stockOrderCarry.userId)
					stockOrderCarry.ResponseWriter.WriteHeader(http.StatusNotFound)
				} else {
					log.Printf("failed to get user stocks for user %s", stockOrderCarry.userId)
					stockOrderCarry.ResponseWriter.WriteHeader(http.StatusInternalServerError)
				}
			}
			continue
		}
		sellerStock := userStocksByPairs[key]

		for _, stockOrderCarry := range stockOrderCarries {
			// Verify seller has the stock and sufficient quantity
			if sellerStock == nil {
				log.Printf("seller does not own stock %s", stockOrderCarry.StockOrder.GetStockIDString())
				stockOrderCarry.ResponseWriter.WriteHeader(http.StatusBadRequest)
				continue
			}
			if sellerStock.UpdateQuantity(-stockOrderCarry.StockOrder.GetQuantity()) != nil {
				log.Printf("insufficient stock quantity: has %d, wants to sell %d\n",
					sellerStock.GetQuantity(), stockOrderCarry.StockOrder.GetQuantity())
				stockOrderCarry.ResponseWriter.WriteHeader(http.StatusBadRequest)
				continue
			}
			stockOrderCarry.UserStock = sellerStock
			_bulkRoutineStockOrderUpdateUserStocks.Insert(stockOrderCarry)
			//matching enigne doesn't care if the user stock is updated or not. It only cares if theres a failure here, and we can cancel the transaction.
			//So we do both at the same time, and just handle things appropriatly.
			_bulkRoutineCreateStockOrderTransactions.Insert(stockOrderCarry)
		}
	}
	return nil
}

func updateUserStocks(data *[]*StockOrderBulk, TransferParams any) error {
	//log.Println("Updating user stocks")
	//map user stocks by id and by stock id
	//then map then map them to the stock orders
	//then we
	userStocks := []userStock.UserStockInterface{}
	for _, stockOrderCarry := range *data {
		userStocks = append(userStocks, stockOrderCarry.UserStock)
	}
	//bulk update user stocks
	//TODO create a setup that errors out only specific parts of the update, not the entire thing.
	errorList, err := _databaseAccessUser.UserStock().UpdateBulk(&userStocks)
	if err != nil {
		log.Printf("Transaction Error failed to update user stocks: %v", err)
		for _, stockOrderCarry := range *data {
			stockOrderCarry.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		}
		return fmt.Errorf("failed to update user stocks: %v", err)
	}

	for _, stockOrderCarry := range *data {
		if errorCode := errorList[stockOrderCarry.UserStock.GetIdString()]; errorCode != 0 {
			log.Println("Stock order with ID: ", stockOrderCarry.UserStock.GetId(), " has Error code: ", errorCode)
			if errorCode == http.StatusNotFound {
				log.Printf("user stock %s not found", stockOrderCarry.UserStock.GetIdString())
				if stockOrderCarry.ResponseWriter.CheckCompleted() {
					cancelStockTransaction(stockOrderCarry.StockOrder.GetIdString())
				} else {
					stockOrderCarry.ResponseWriter.WriteHeader(http.StatusNotFound)
				}
			} else {
				log.Printf("failed to update user stock %s", stockOrderCarry.UserStock.GetIdString())
				if stockOrderCarry.ResponseWriter.CheckCompleted() {
					cancelStockTransaction(stockOrderCarry.StockOrder.GetIdString())
				} else {
					stockOrderCarry.ResponseWriter.WriteHeader(http.StatusInternalServerError)
				}
			}
		}
		if stockOrderCarry.ResponseWriter.CheckCompleted() && stockOrderCarry.ResponseWriter.GetStatusCode() != http.StatusOK {
			stockOrderCarry.UserStock.UpdateQuantity(stockOrderCarry.StockOrder.GetQuantity())
			_bulkRoutineStockOrderUpdateUserStocks.Insert(stockOrderCarry)
		}
	}
	return nil
}

func placeStockOrderResponse(data *[]*StockOrderBulk, TransferParams any) error {
	//log.Println("Creating stock order transactions")
	bulkTransactions := make([]transaction.StockTransactionInterface, len(*data))
	i := 0
	for _, stockOrderCarry := range *data {
		if !stockOrderCarry.ResponseWriter.CheckCompleted() {
			newTransaction := transaction.NewStockTransaction(transaction.NewStockTransactionParams{
				StockOrder:  stockOrderCarry.StockOrder,
				OrderStatus: "IN_PROGRESS",
			})
			newTransaction.SetStockID(stockOrderCarry.StockOrder.GetStockID())
			newTransaction.SetUnqiuePairing(stockOrderCarry.StockOrder.GetUniquePairing())
			bulkTransactions[i] = newTransaction
			i++
		}
	}
	createdTransactions, errList, err := _databaseAccess.StockTransaction().CreateBulk(&bulkTransactions)
	if err != nil {
		for _, stockOrderCarry := range *data {
			//only way this could be registered as completed is if updating the user stocks ran into an error.
			//otherwise we consider that it has not completed, and we need to undo that update to the quantity.
			if !stockOrderCarry.ResponseWriter.CheckCompleted() {
				stockOrderCarry.ResponseWriter.WriteHeader(http.StatusInternalServerError)
				if stockOrderCarry.UserStock != nil {
					stockOrderCarry.UserStock.UpdateQuantity(stockOrderCarry.StockOrder.GetQuantity())
					_bulkRoutineStockOrderUpdateUserStocks.Insert(stockOrderCarry)
				}
			}
		}
		log.Printf("failed to create transactions: %v", err)
		return fmt.Errorf("failed to create transactions: %v", err)
	}
	createdTransactionIdsByPairing := make(map[string]*uuid.UUID)
	for _, transaction := range *createdTransactions {
		createdTransactionIdsByPairing[transaction.GetUniquePairing().String()] = transaction.GetId()
	}

	for _, stockOrderCarry := range *data {
		if stockOrderCarry.ResponseWriter.CheckCompleted() {
			err := _networkHttpManager.Transactions().Patch("cancelStockTransaction", stockOrderCarry.StockOrder.GetIdString())
			if err != nil {
				log.Println("Error: ", err.Error())
				// panic(err)
			}
			continue
		}
		if val, ok := errList[stockOrderCarry.StockOrder.GetUniquePairing().String()]; ok && val != 0 {
			//only way this could be registered as completed is if updating the user stocks ran into an error.
			//otherwise we consider that it has not completed, and we need to undo that update to the quantity.
			if !stockOrderCarry.ResponseWriter.CheckCompleted() {
				stockOrderCarry.ResponseWriter.WriteHeader(http.StatusInternalServerError)
				if stockOrderCarry.UserStock != nil {
					stockOrderCarry.UserStock.UpdateQuantity(stockOrderCarry.StockOrder.GetQuantity())
					_bulkRoutineStockOrderUpdateUserStocks.Insert(stockOrderCarry)
				}
			}
			continue
		}
		stockOrderCarry.StockOrder.SetId(createdTransactionIdsByPairing[stockOrderCarry.StockOrder.GetUniquePairing().String()])
		//log.Println("sending to matching engine")
		_, err = _networkQueueManager.MatchingEngine().Post("placeStockOrder", stockOrderCarry.StockOrder)
		//log.Println("sent to matching engine")
		if err != nil {
			log.Printf("failed to send to matching engine: %v", err)
			err = cancelStockTransaction(stockOrderCarry.StockOrder.GetIdString())
			if err != nil {
				log.Println("Error: ", err.Error())
				// panic(err)
			}
			//only way this could be registered as completed is if updating the user stocks ran into an error.
			//otherwise we consider that it has not completed, and we need to undo that update to the quantity.
			if !stockOrderCarry.ResponseWriter.CheckCompleted() {
				stockOrderCarry.ResponseWriter.WriteHeader(http.StatusInternalServerError)
				if stockOrderCarry.UserStock != nil {
					stockOrderCarry.UserStock.UpdateQuantity(stockOrderCarry.StockOrder.GetQuantity())
					_bulkRoutineStockOrderUpdateUserStocks.Insert(stockOrderCarry)
				}
			}
			continue
		}
		if stockOrderCarry.ResponseWriter.CheckCompleted() {
			err = cancelStockTransaction(stockOrderCarry.StockOrder.GetIdString())
			if err != nil {
				log.Println("Error: ", err.Error())
				// panic(err)
			}
			continue
		}
		returnVal := network.ReturnJSON{
			Success: true,
			Data:    nil,
		}
		returnValJSON, err := json.Marshal(returnVal)
		if err != nil {
			log.Printf("failed to marshal return value: %v", err)
			if !stockOrderCarry.ResponseWriter.CheckCompleted() {
				stockOrderCarry.ResponseWriter.WriteHeader(http.StatusInternalServerError)
			}
			continue
		}
		//log.Println("value return")
		stockOrderCarry.ResponseWriter.Write(returnValJSON)
	}
	return nil
}

func cancelStockTransactionHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	//log.Println("Cancelling stock transaction")
	var stockTxID network.StockTransactionID
	err := json.Unmarshal(data, &stockTxID)
	if err != nil {
		log.Println("Error: ", err.Error())
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	err = cancelStockTransaction(stockTxID.StockTransactionID)

	if err != nil {
		log.Println("Error: ", err.Error())
		if err.Error() == "404 Not Found" {
			responseWriter.WriteHeader(http.StatusNotFound)
		} else {
			responseWriter.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	returnVal := network.ReturnJSON{
		Success: true,
		Data:    nil,
	}
	returnValJSON, err := json.Marshal(returnVal)
	if err != nil {
		log.Println("Error: ", err.Error())
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

	_, err = _networkHttpManager.MatchingEngine().Delete("deleteOrder/" + id)
	if err != nil {
		log.Println("Error: ", err.Error())
		return err
	}
	return nil

}

// func placeStockOrderHandlerOld(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
// 	//println("Placing stock order")
// 	stockOrder, err := order.Parse(data)
// 	if err != nil {
// 		println("Error: ", err.Error())
// 		responseWriter.WriteHeader(http.StatusBadRequest)
// 		return
// 	}
// 	uuidNew, err := uuid.Parse(strings.TrimSpace(queryParams.Get("userID")))
// 	if err != nil {
// 		println("Error: ", err.Error())
// 		responseWriter.WriteHeader(http.StatusBadRequest)
// 		return
// 	}
// 	stockOrder.SetUserID(&uuidNew)
// 	err = placeStockOrderOld(stockOrder)
// 	if err != nil {
// 		println("Error: ", err.Error())
// 		responseWriter.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}
// 	returnVal := network.ReturnJSON{
// 		Success: true,
// 		Data:    nil,
// 	}
// 	returnValJSON, err := json.Marshal(returnVal)
// 	if err != nil {
// 		responseWriter.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}
// 	responseWriter.Write(returnValJSON)
// }

// func placeStockOrderOld(stockOrder order.StockOrderInterface) error {
// 	var err error

// 	if !stockOrder.GetIsBuy() {
// 		// Get seller's current stock holdings
// 		sellerStockPortfolio, err := _databaseAccessUser.UserStock().GetUserStocks(stockOrder.GetUserIDString())
// 		if err != nil {
// 			return fmt.Errorf("failed to get seller stocks: %v", err)
// 		}

// 		// Find the stock in the seller's portfolio
// 		var sellerStock userStock.UserStockInterface
// 		for _, stock := range *sellerStockPortfolio {
// 			if stock.GetStockIDString() == stockOrder.GetStockIDString() {
// 				sellerStock = stock
// 				break
// 			}
// 		}

// 		// Verify seller has the stock and sufficient quantity
// 		if sellerStock == nil {
// 			return fmt.Errorf("seller does not own stock %s", stockOrder.GetStockID())
// 		}
// 		if sellerStock.GetQuantity() < stockOrder.GetQuantity() {
// 			return fmt.Errorf("insufficient stock quantity: has %d, wants to sell %d",
// 				sellerStock.GetQuantity(), stockOrder.GetQuantity())
// 		}

// 		// Deduct the quantity from seller's portfolio but keep the record
// 		sellerStock.UpdateQuantity(-stockOrder.GetQuantity())
// 		err = _databaseAccessUser.UserStock().Update(sellerStock)
// 		if err != nil {
// 			return fmt.Errorf("failed to update seller stock quantity: %v", err)
// 		}
// 	}

// 	transaction := transaction.NewStockTransaction(transaction.NewStockTransactionParams{
// 		StockOrder:  stockOrder,
// 		OrderStatus: "IN_PROGRESS",
// 	})

// 	createdTransaction, err := _databaseAccess.StockTransaction().Create(transaction)
// 	if err != nil {
// 		println("Error: ", err.Error())
// 		return err
// 	}
// 	stockOrder.SetId(createdTransaction.GetId())
// 	//pass to matching engine
// 	_, err = _networkHttpManager.MatchingEngine().Post("placeStockOrder", stockOrder)
// 	return err
// }
