package matchingEngine

import (
	"Shared/entities/order"
	"Shared/entities/transaction"
	"Shared/network"
	"databaseAccessStockOrder"
	"encoding/json"
	"net/http"
)

var _matchingEngineMap map[string]MatchingEngineInterface
var _databaseManager databaseAccessStockOrder.DatabaseAccessInterface
var _networkManager network.HttpClientInterface

func InitalizeHandlers(stockIDs *[]string,
	networkManager network.HttpClientInterface, databaseManager databaseAccessStockOrder.DatabaseAccessInterface) {
	_databaseManager = databaseManager
	_networkManager = networkManager
	_matchingEngineMap = make(map[string]MatchingEngineInterface)
	//Create all matching engines for stocks.
	for _, stockID := range *stockIDs {
		AddNewStock(stockID)
	}

	//Add handlers
	networkManager.AddHandleFunc(network.HandlerParams{Pattern: "/createStock", Handler: AddNewStockHandler})
	networkManager.AddHandleFunc(network.HandlerParams{Pattern: "/placeOrder", Handler: PlaceStockOrderHandler})
	networkManager.AddHandleFunc(network.HandlerParams{Pattern: "/deleteOrder", Handler: DeleteStockOrderHandler})
}

// Expected input is a stock ID in the body of the request
// we're expecting {"StockID":"{id value}"}
func AddNewStockHandler(responseWriter http.ResponseWriter, data []byte) {
	var jsonData map[string]interface{}
	err := json.Unmarshal(data, &jsonData)
	if err != nil {
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	stockID, ok := jsonData["StockID"].(string)
	if !ok {
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	AddNewStock(stockID)
}

func AddNewStock(stockID string) {
	_, ok := _matchingEngineMap[stockID]
	//if we don't have a matching engine for this stock, create one
	if !ok {
		stockOrders := _databaseManager.GetInitialStockOrdersForStock(stockID)
		ordersInterface := make([]order.StockOrderInterface, len(*stockOrders))
		copy(ordersInterface, *stockOrders)
		me := NewMatchingEngineForStock(&NewMatchingEngineParams{
			StockID:                  stockID,
			InitalOrders:             &ordersInterface,
			SendToOrderExecutionFunc: SendToOrderExection,
			DatabaseManager:          _databaseManager,
		})
		_matchingEngineMap[stockID] = me
		go me.RunMatchingEngineOrders()
		go me.RunMatchingEngineUpdates()
	}
}

func PlaceStockOrderHandler(responseWriter http.ResponseWriter, data []byte) {
	//parse the stock order
	stockOrder, err := order.Parse(data)
	if err != nil {
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	if PlaceStockOrder(stockOrder) {
		responseWriter.WriteHeader(http.StatusOK)
	} else {
		responseWriter.WriteHeader(http.StatusBadRequest)
	}
}

func PlaceStockOrder(stockOrder order.StockOrderInterface) bool {
	me, ok := _matchingEngineMap[stockOrder.GetStockID()]
	if !ok {
		return false
	}

	stockOrderCreated := _databaseManager.CreateStockOrder(stockOrder)
	me.AddOrder(stockOrderCreated)
	return true
}

func DeleteStockOrderHandler(responseWriter http.ResponseWriter, data []byte) {
	var jsonData map[string]interface{}
	err := json.Unmarshal(data, &jsonData)
	if err != nil {
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	orderID, ok := jsonData["OrderID"].(string)
	if !ok {
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	DeleteStockOrder(orderID)
}

func DeleteStockOrder(orderID string) {
	stockOrder := (_databaseManager.DeleteStockOrder(orderID))
	if stockOrder == nil {
		return
	}
	me, ok := _matchingEngineMap[orderID]
	if !ok {
		return
	}
	me.RemoveOrder(stockOrder.GetId(), stockOrder.GetPrice())
}

func SendToOrderExection(buyOrder order.StockOrderInterface, sellOrder order.StockOrderInterface, childOrder order.StockOrderInterface) transaction.StockTransactionInterface {
	var childQuantityBuying float64
	var childQuantitySelling float64
	if buyOrder.GetQuantity() >= sellOrder.GetQuantity() {
		childQuantityBuying = sellOrder.GetQuantity()
	}
	if buyOrder.GetQuantity() <= sellOrder.GetQuantity() {
		childQuantitySelling = buyOrder.GetQuantity()
	}

	data, err := _networkManager.Post("???", network.MatchingEngineToExectuionJSON{
		StockID:         buyOrder.GetStockID(),
		BuyOrderID:      buyOrder.GetId(),
		SellOrderID:     sellOrder.GetId(),
		StockPrice:      sellOrder.GetPrice(),
		FullQuantityBuying:  buyOrder.GetQuantity(),
		ThisQuantityBuying:  ,
		FullQuantitySelling: sellOrder.GetQuantity(),
		ThisQuantitySelling: childOrder.GetQuantity(),
	})
	if err != nil {
		return nil
	}
	transaction, errParse := transaction.ParseStockTransaction(data)
	if errParse != nil {
		return nil
	}

	//send to order execution
	return transaction
}
