package matchingEngine

import (
	"Shared/entities/order"
	"Shared/network"
	subfunctions "Shared/subfunctions/Multithreading"
	"databaseAccessStockOrder"
	"databaseAccessUserManagement"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var _matchingEngineMap map[string]MatchingEngineInterface
var _databaseManagerStockOrder databaseAccessStockOrder.DatabaseAccessInterface
var _networkHttpManager network.NetworkInterface
var _networkQueueManager network.NetworkInterface
var _databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface

var _bulkStockOrderAdder subfunctions.BulkRoutineInterface[*StockOrderBulk]

type StockOrderBulk struct {
	StockOrder     order.StockOrderInterface
	ResponseWriter network.ResponseWriter
}

var stockPriceIndex []string
var stockIdToName map[string]string

func InitalizeHandlers(stockIDs *[]network.StockPrice,
	networkHttpManager network.NetworkInterface, networkQueueManager network.NetworkInterface, databaseManagerStockOrder databaseAccessStockOrder.DatabaseAccessInterface, databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface) {
	_databaseManagerStockOrder = databaseManagerStockOrder
	_databaseAccessUser = databaseAccessUser
	_networkHttpManager = networkHttpManager
	_networkQueueManager = networkQueueManager
	_matchingEngineMap = make(map[string]MatchingEngineInterface)
	stockPriceIndex = make([]string, 0)
	stockIdToName = make(map[string]string)

	//Create all matching engines for stocks.
	for _, stockID := range *stockIDs {
		AddNewStock(stockID.StockID, stockID.StockName)
	}

	_bulkStockOrderAdder = subfunctions.NewBulkRoutine[*StockOrderBulk](&subfunctions.BulkRoutineParams[*StockOrderBulk]{
		Routine: PlaceStockOrder,
	})

	//Add handlers
	_networkQueueManager.AddHandleFuncUnprotected(network.HandlerParams{Pattern: "createStock", Handler: AddNewStockHandler})
	_networkQueueManager.AddHandleFuncUnprotected(network.HandlerParams{Pattern: "placeStockOrder", Handler: PlaceStockOrderHandler})
	_networkQueueManager.AddHandleFuncUnprotected(network.HandlerParams{Pattern: "deleteOrder/", Handler: DeleteStockOrderHandler})
	_networkHttpManager.AddHandleFuncProtected(network.HandlerParams{Pattern: os.Getenv("transaction_route") + "/getStockPrices", Handler: GetStockPricesHandler})
	_networkQueueManager.AddHandleFuncUnprotected(network.HandlerParams{Pattern: "CompletePairedOrder", Handler: CompletePairedOrderHandler})
	http.HandleFunc("/health", healthHandler)
	networkQueueManager.Listen()
	//Add a new queue listener to handle the confirmation of database updates on stock orders and transactions.
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Simple check: you might expand this to test database connectivity, etc.
	w.WriteHeader(http.StatusOK)
	//log.Println(w, "OK")
}

// Expected input is a stock ID in the body of the request
// we're expecting {"StockID":"{id value}"}
func AddNewStockHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	var stockID network.StockID
	err := json.Unmarshal(data, &stockID)
	if err != nil {
		log.Println("Error: ", err.Error())
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	// uid, err := uuid.Parse(stockID.StockID)
	// if err != nil {
	// 	log.Println("Error: ", err.Error())
	// 	responseWriter.WriteHeader(http.StatusBadRequest)
	// 	return
	// }
	AddNewStock(stockID.StockID, stockID.Name)
	responseWriter.WriteHeader(http.StatusOK)
}

func AddNewStock(stockID string, stockName string) {
	_, ok := _matchingEngineMap[stockID]
	//if we don't have a matching engine for this stock, create one
	if !ok {
		stockOrders := _databaseManagerStockOrder.GetInitialStockOrdersForStock(stockID)
		ordersInterface := make([]order.StockOrderInterface, len(*stockOrders))
		copy(ordersInterface, *stockOrders)

		me := NewMatchingEngineForStock(&NewMatchingEngineParams{
			StockID:                  stockID,
			InitalOrders:             &ordersInterface,
			SendToOrderExecutionFunc: SendToOrderExection,
			DatabaseManager:          _databaseManagerStockOrder,
			StockName:                stockName,
		})
		_matchingEngineMap[stockID] = me
		go me.RunMatchingEngineOrders()
		go me.RunMatchingEngineUpdates()
		stockIdToName[stockID] = stockName
		stockPriceIndex = append(stockPriceIndex, stockID)
		//sort these here so that we don't have to sort them in every single price call.
		sort.SliceStable(stockPriceIndex, func(i, j int) bool {
			return stockIdToName[stockPriceIndex[i]] > stockIdToName[stockPriceIndex[j]]
		})
	}
}

func PlaceStockOrderHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	//parse the stock order
	stockOrder, err := order.Parse(data)
	if err != nil {
		log.Println("Error: ", err.Error())
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	_bulkStockOrderAdder.Insert(&StockOrderBulk{
		StockOrder:     stockOrder,
		ResponseWriter: responseWriter,
	})
	// if PlaceStockOrderOld(stockOrder) {
	// 	responseWriter.WriteHeader(http.StatusOK)
	// } else {
	// 	responseWriter.WriteHeader(http.StatusBadRequest)
	// }
}

func PlaceStockOrderOld(stockOrder order.StockOrderInterface) bool {
	if me, ok := _matchingEngineMap[stockOrder.GetStockID().String()]; ok {
		createdOrder, err := _databaseManagerStockOrder.Create(stockOrder)
		if err != nil {
			log.Println("Error: ", err.Error())
			return false
		}
		me.AddOrder(createdOrder)
		return true
	}
	log.Println("Error: Matching engine not found for ID: ", stockOrder.GetStockIDString())
	return false
}

func PlaceStockOrder(data *[]*StockOrderBulk, TransferParams any) error {
	stockOrderPairings := make(map[string]*StockOrderBulk, len(*data))
	stockOrderList := make([]order.StockOrderInterface, len(*data))
	for i, stockOrderBulk := range *data {
		if _, ok := _matchingEngineMap[stockOrderBulk.StockOrder.GetStockIDString()]; !ok {
			log.Println("Error: Matching engine not found for ID: ", stockOrderBulk.StockOrder.GetStockIDString(), " / ", stockOrderBulk.StockOrder.GetStockID())
			stockOrderBulk.ResponseWriter.WriteHeader(http.StatusBadRequest)
			continue
		}
		stockOrderPairings[stockOrderBulk.StockOrder.GetUniquePairing().String()] = stockOrderBulk
		stockOrderList[i] = stockOrderBulk.StockOrder
	}
	if len(stockOrderList) == 0 {
		log.Println("No stock orders to place")
		return nil
	}
	go func() {
		var errorList map[string]int
		stockOrders, errorList, err := _databaseManagerStockOrder.CreateBulk(&stockOrderList)
		if err != nil {
			log.Println("Error: ", err.Error())
			for _, stockOrderBulk := range stockOrderPairings {
				stockOrderBulk.ResponseWriter.WriteHeader(http.StatusInternalServerError)
			}
		}
		for _, stockOrder := range *stockOrders {
			if _, ok := errorList[stockOrder.GetUniquePairing().String()]; ok {
				_matchingEngineMap[stockOrder.GetStockIDString()].RemoveOrder(stockOrder.GetIdString(), stockOrder.GetPrice())
				stockOrderPairings[stockOrder.GetUniquePairing().String()].ResponseWriter.WriteHeader(http.StatusBadRequest)
				continue
			}
		}
	}()
	for _, stockOrder := range stockOrderList {
		if me, ok := _matchingEngineMap[stockOrder.GetStockIDString()]; ok {
			me.AddOrder(stockOrder)
			stockOrderPairings[stockOrder.GetUniquePairing().String()].ResponseWriter.WriteHeader(http.StatusOK)
		} else {
			log.Println("Error: Matching engine not found for ID: ", stockOrder.GetStockIDString())
			stockOrderPairings[stockOrder.GetUniquePairing().String()].ResponseWriter.WriteHeader(http.StatusBadRequest)
		}
	}
	return nil
}

// this handles the delete order request from initiator, spawned by /cancelStockOrder from user
func DeleteStockOrderHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	orderID, err := uuid.Parse(strings.TrimSpace(queryParams.Get("id")))
	if err != nil {
		log.Println("Error: ", err.Error())
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	err = DeleteStockOrder(&orderID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		responseWriter.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		log.Println("Error: ", err.Error())
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	responseWriter.WriteHeader(http.StatusOK)
}

func DeleteStockOrder(orderID *uuid.UUID) error {
	order, err := _databaseManagerStockOrder.GetByID(orderID)
	if err != nil {
		log.Println("Cancel Stock GetByID Error: ", err.Error())
		return err
	}
	err = _databaseManagerStockOrder.Delete(orderID)
	if err != nil {
		log.Println("Cancel Stock Delete Error: ", err.Error())
		return err
	}
	me, ok := _matchingEngineMap[order.GetStockIDString()]
	if !ok {
		log.Println("Cancel Stock Error: Matching engine not found for ID: ", order.GetStockID())
		return nil
	}
	me.RemoveOrder(orderID.String(), order.GetPrice())
	return nil
}

func GetStockPricesHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	stockPrices := make([]network.StockPrice, len(stockPriceIndex))
	// log.Println("Stock Price Index length: ", len(stockPriceIndex))
	for i, stockID := range stockPriceIndex {
		// val := _matchingEngineMap[stockID].GetPrice()
		// log.Println(i, ". Appending stock: "+val.StockName, " with price: ", val.Price)
		stockPrices[i] = _matchingEngineMap[stockID].GetPrice()
	}
	returnVal := network.ReturnJSON{
		Success: true,
		Data:    stockPrices,
	}
	pricesJSON, err := json.Marshal(returnVal)
	if err != nil {
		log.Println("Error: ", err.Error())
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	responseWriter.Write(pricesJSON)
}

func SendToOrderExection(buyOrder order.StockOrderInterface, sellOrder order.StockOrderInterface) (network.ExecutorToMatchingEngineJSON, error) {
	buyQty := buyOrder.GetQuantity()
	sellQty := sellOrder.GetQuantity()
	quantity := buyQty
	if sellQty < buyQty {
		quantity = sellQty
	}
	// transferEntity := network.MatchingEngineToExecutionJSON{
	// 	BuyerID:     buyOrder.GetUserIDString(),
	// 	SellerID:    sellOrder.GetUserIDString(),
	// 	StockID:     buyOrder.GetStockIDString(),
	// 	BuyOrderID:  buyOrder.GetIdString(),
	// 	SellOrderID: sellOrder.GetIdString(),
	// 	StockPrice:  sellOrder.GetPrice(),
	// 	Quantity:    quantity,
	// 	Name:        stockIdToName[buyOrder.GetStockIDString()],
	// }

	data, err := _databaseAccessUser.ExecuteOrder(buyOrder.GetIdString(), sellOrder.GetIdString(), buyOrder.GetUserIDString(), sellOrder.GetUserIDString(), buyOrder.GetStockIDString(), stockIdToName[buyOrder.GetStockIDString()], sellOrder.GetPrice(), quantity)

	// if err != nil {
	// 	log.Println("Error: ", err.Error())
	// 	return network.ExecutorToMatchingEngineJSON{
	// 		StockID:            buyOrder.GetStockIDString(),
	// 		BuyerStockOrderId:  buyOrder.GetIdString(),
	// 		SellerStockOrderId: sellOrder.GetIdString(),
	// 		IsBuyFailure:       true,
	// 		IsSellFailure:      true,
	// 	}, err
	// }
	//var matchedData network.ExecutorToMatchingEngineJSON
	//err = json.Unmarshal(data, &matchedData)
	// if err != nil {
	// 	log.Println("Error: ", err.Error())
	// 	return network.ExecutorToMatchingEngineJSON{
	// 		StockID:            buyOrder.GetStockIDString(),
	// 		BuyerStockOrderId:  buyOrder.GetIdString(),
	// 		SellerStockOrderId: sellOrder.GetIdString(),
	// 		IsBuyFailure:       true,
	// 		IsSellFailure:      true,
	// 	}, err
	// }
	return data, err
}

func CompletePairedOrderHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	//parse the stock order
	var matchedData network.ExecutorToMatchingEngineJSON
	err := json.Unmarshal(data, &matchedData)
	if err != nil {
		log.Println("Error: ", err.Error())
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	me, ok := _matchingEngineMap[matchedData.StockID]
	if !ok {
		log.Println("Error: Matching engine not found for ID: ", matchedData.StockID)
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	err = me.CompletePairedOrder(matchedData)
	if err != nil {
		log.Println("Error: ", err.Error())
		//check if the error string has 404
		if strings.Contains(err.Error(), "404") {
			responseWriter.WriteHeader(http.StatusNotFound)
			return
		} else {
			responseWriter.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
