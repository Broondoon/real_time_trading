package matchingEngine

import (
	"Shared/entities/order"
	"Shared/network"
	subfunctions "Shared/subfunctions/Multithreading"
	"databaseAccessStock"
	"databaseAccessStockOrder"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var _matchingEngineMap map[string]MatchingEngineInterface
var _databaseManager databaseAccessStockOrder.DatabaseAccessInterface
var _networkHttpManager network.NetworkInterface
var _networkQueueManager network.NetworkInterface
var _stockDatabaseAccess databaseAccessStock.DatabaseAccessInterface

var _bulkStockOrderAdder subfunctions.BulkRoutineInterface[*StockOrderBulk]

type StockOrderBulk struct {
	StockOrder     order.StockOrderInterface
	ResponseWriter network.ResponseWriter
}

var stockPriceIndex map[string]int
var stockPrices []network.StockPrice

var updatePrice chan network.StockPrice

func InitalizeHandlers(stockIDs *[]network.StockPrice,
	networkHttpManager network.NetworkInterface, networkQueueManager network.NetworkInterface, databaseManager databaseAccessStockOrder.DatabaseAccessInterface, stockDatabaseAccess databaseAccessStock.DatabaseAccessInterface) {
	_databaseManager = databaseManager
	_networkHttpManager = networkHttpManager
	_networkQueueManager = networkQueueManager
	_stockDatabaseAccess = stockDatabaseAccess
	_matchingEngineMap = make(map[string]MatchingEngineInterface)
	stockPriceIndex = make(map[string]int)
	stockPrices = make([]network.StockPrice, len(*stockIDs))
	updatePrice = make(chan network.StockPrice, 1000)

	//Create all matching engines for stocks.
	for _, stockID := range *stockIDs {
		AddNewStock(stockID.StockID, stockID.StockName)
	}

	_bulkStockOrderAdder = subfunctions.NewBulkRoutine[*StockOrderBulk](&subfunctions.BulkRoutineParams[*StockOrderBulk]{
		Routine: PlaceStockOrder,
	})

	//Add handlers
	_networkHttpManager.AddHandleFuncUnprotected(network.HandlerParams{Pattern: "createStock", Handler: AddNewStockHandler})
	_networkHttpManager.AddHandleFuncUnprotected(network.HandlerParams{Pattern: "placeStockOrder", Handler: PlaceStockOrderHandler})
	_networkHttpManager.AddHandleFuncUnprotected(network.HandlerParams{Pattern: "deleteOrder/", Handler: DeleteStockOrderHandler})
	_networkHttpManager.AddHandleFuncProtected(network.HandlerParams{Pattern: os.Getenv("transaction_route") + "/getStockPrices", Handler: GetStockPricesHandler})
	http.HandleFunc("/health", healthHandler)

	go func() {
		for price := range updatePrice {
			stockPrices[stockPriceIndex[price.StockID]].Price = price.Price
		}
	}()

	networkQueueManager.Listen()
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
		stockOrders := _databaseManager.GetInitialStockOrdersForStock(stockID)
		ordersInterface := make([]order.StockOrderInterface, len(*stockOrders))
		copy(ordersInterface, *stockOrders)
		stockPrices = append(stockPrices, network.StockPrice{
			StockID:   stockID,
			StockName: stockName,
			Price:     0,
		})
		sort.SliceStable(stockPrices, func(i, j int) bool {
			return stockPrices[i].StockName > stockPrices[j].StockName
		})
		for i, stockPrice := range stockPrices {
			stockPriceIndex[stockPrice.StockID] = i
		}
		me := NewMatchingEngineForStock(&NewMatchingEngineParams{
			StockID:                  stockID,
			InitalOrders:             &ordersInterface,
			SendToOrderExecutionFunc: SendToOrderExection,
			DatabaseManager:          _databaseManager,
			UpdatePrice:              &updatePrice,
		})
		_matchingEngineMap[stockID] = me
		go me.RunMatchingEngineOrders()
		go me.RunMatchingEngineUpdates()
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
		createdOrder, err := _databaseManager.Create(stockOrder)
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
	var errors map[string]int
	log.Println("Stock Order List len: ", len(stockOrderList))
	stockOrders, errors, err := _databaseManager.CreateBulk(&stockOrderList)
	if err != nil {
		log.Println("Error: ", err.Error())
		for _, stockOrderBulk := range stockOrderPairings {
			stockOrderBulk.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		}
	}
	for _, stockOrder := range *stockOrders {
		if _, ok := errors[stockOrder.GetUniquePairing().String()]; ok {
			stockOrderPairings[stockOrder.GetUniquePairing().String()].ResponseWriter.WriteHeader(http.StatusBadRequest)
			continue
		}
		me := _matchingEngineMap[stockOrder.GetStockIDString()]
		me.AddOrder(stockOrder)
		stockOrderPairings[stockOrder.GetUniquePairing().String()].ResponseWriter.WriteHeader(http.StatusOK)
	}
	return nil
}

func DeleteStockOrderHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	orderID, err := uuid.Parse(queryParams.Get("id"))
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
	order, err := _databaseManager.GetByID(orderID)
	if err != nil {
		log.Println("Error: ", err.Error())
		return err
	}
	err = _databaseManager.Delete(orderID)
	if err != nil {
		log.Println("Error: ", err.Error())
		return err
	}
	me, ok := _matchingEngineMap[order.GetStockIDString()]
	if !ok {
		log.Println("Error: Matching engine not found for ID: ", order.GetStockID())
		return nil
	}
	me.RemoveOrder(orderID.String(), order.GetPrice())
	return nil
}

func GetStockPricesHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {

	returnVal := network.ReturnJSON{
		Success: true,
		Data:    &stockPrices,
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
	transferEntity := network.MatchingEngineToExecutionJSON{
		BuyerID:       buyOrder.GetUserIDString(),
		SellerID:      sellOrder.GetUserIDString(),
		StockID:       buyOrder.GetStockIDString(),
		BuyOrderID:    buyOrder.GetIdString(),
		SellOrderID:   sellOrder.GetIdString(),
		IsBuyPartial:  buyQty > sellQty,
		IsSellPartial: buyQty < sellQty,
		StockPrice:    sellOrder.GetPrice(),
		Quantity:      quantity,
		Name:          stockPrices[stockPriceIndex[buyOrder.GetStockIDString()]].StockName,
	}

	data, err := _networkQueueManager.OrderExecutor().Post("executor", transferEntity)

	if err != nil {
		log.Println("Error: ", err.Error())
		return network.ExecutorToMatchingEngineJSON{}, err
	}
	var matchedData network.ExecutorToMatchingEngineJSON
	// matchedData = network.ExecutorToMatchingEngineJSON{
	// 	IsBuyFailure:  false,
	// 	IsSellFailure: false,
	// }
	err = json.Unmarshal(data, &matchedData)
	if err != nil {
		log.Println("Error: ", err.Error())
		return network.ExecutorToMatchingEngineJSON{}, err
	}
	return matchedData, nil
}
