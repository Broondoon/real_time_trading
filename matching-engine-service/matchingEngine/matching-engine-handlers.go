package matchingEngine

import (
	"Shared/entities/order"
	"Shared/network"
	"databaseAccessStock"
	"databaseAccessStockOrder"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"sort"

	"gorm.io/gorm"
)

var _matchingEngineMap map[string]MatchingEngineInterface
var _databaseManager databaseAccessStockOrder.DatabaseAccessInterface
var _networkHttpManager network.NetworkInterface
var _networkQueueManager network.NetworkInterface
var _stockDatabaseAccess databaseAccessStock.DatabaseAccessInterface

func InitalizeHandlers(stockIDs *[]string,
	networkHttpManager network.NetworkInterface, networkQueueManager network.NetworkInterface, databaseManager databaseAccessStockOrder.DatabaseAccessInterface, stockDatabaseAccess databaseAccessStock.DatabaseAccessInterface) {
	_databaseManager = databaseManager
	_networkHttpManager = networkHttpManager
	_networkQueueManager = networkQueueManager
	_stockDatabaseAccess = stockDatabaseAccess
	_matchingEngineMap = make(map[string]MatchingEngineInterface)
	//Create all matching engines for stocks.
	for _, stockID := range *stockIDs {
		AddNewStock(stockID)
	}

	//Add handlers
	_networkHttpManager.AddHandleFuncUnprotected(network.HandlerParams{Pattern: "createStock", Handler: AddNewStockHandler})
	_networkQueueManager.AddHandleFuncUnprotected(network.HandlerParams{Pattern: "placeStockOrder", Handler: PlaceStockOrderHandler})
	_networkQueueManager.AddHandleFuncUnprotected(network.HandlerParams{Pattern: "deleteOrder/", Handler: DeleteStockOrderHandler})
	_networkHttpManager.AddHandleFuncProtected(network.HandlerParams{Pattern: os.Getenv("transaction_route") + "/getStockPrices", Handler: GetStockPricesHandler})
	http.HandleFunc("/health", healthHandler)
	networkQueueManager.Listen()
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Simple check: you might expand this to test database connectivity, etc.
	w.WriteHeader(http.StatusOK)
	//fmt.Println(w, "OK")
}

// Expected input is a stock ID in the body of the request
// we're expecting {"StockID":"{id value}"}
func AddNewStockHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	println("Adding new stock")
	println("Data: ", string(data))
	println("Query Params: ", queryParams.Encode())
	println("Request Type: ", requestType)
	var stockID network.StockID
	err := json.Unmarshal(data, &stockID)
	if err != nil {
		println("Error: ", err.Error())
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	AddNewStock(stockID.StockID)
	responseWriter.WriteHeader(http.StatusOK)
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

func PlaceStockOrderHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	println("Received stock order")
	println("Data: ", string(data))
	//parse the stock order
	stockOrder, err := order.Parse(data)
	if err != nil {
		println("Error: ", err.Error())
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
	println("Placing stock order")
	if me, ok := _matchingEngineMap[stockOrder.GetStockID()]; ok {
		createdOrder, err := _databaseManager.Create(stockOrder)
		if err != nil {
			println("Error: ", err.Error())
			return false
		}
		me.AddOrder(createdOrder)
		return true
	}
	println("Error: Matching engine not found for ID: ", stockOrder.GetStockID())
	return false
}

func DeleteStockOrderHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	orderID := queryParams.Get("id")
	err := DeleteStockOrder(orderID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		responseWriter.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		println("Error: ", err.Error())
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	responseWriter.WriteHeader(http.StatusOK)
}

func DeleteStockOrder(orderID string) error {
	order, err := _databaseManager.GetByID(orderID)
	if err != nil {
		println("Error: ", err.Error())
		return err
	}
	err = _databaseManager.Delete(orderID)
	if err != nil {
		println("Error: ", err.Error())
		return err
	}
	me, ok := _matchingEngineMap[order.GetStockID()]
	if !ok {
		println("Error: Matching engine not found for ID: ", order.GetStockID())
		return nil
	}
	me.RemoveOrder(orderID, order.GetPrice())
	return nil
}

func GetStockPricesHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	println("Getting stock prices")
	prices, err := GetStockPrices()
	if err != nil {
		println("Error: ", err.Error())
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	returnVal := network.ReturnJSON{
		Success: true,
		Data:    prices,
	}
	pricesJSON, err := json.Marshal(returnVal)
	if err != nil {
		println("Error: ", err.Error())
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	responseWriter.Write(pricesJSON)
}

func GetStockPrices() (*[]network.StockPrice, error) {
	stocks, err := _stockDatabaseAccess.GetAll()
	if err != nil {
		println("Error: ", err.Error())
		return nil, err
	}
	//create a map from the stock ids to names
	stockIDToName := make(map[string]string)
	for _, stock := range *stocks {
		stockIDToName[stock.GetId()] = stock.GetName()
	}
	//get the prices for each stock
	prices := make(map[string]float64)
	for stockID, me := range _matchingEngineMap {
		prices[stockID] = me.GetPrice()
	}
	//create the stock prices
	stockPrices := make([]network.StockPrice, len(prices))
	i := 0
	for stockID, price := range prices {
		stockPrices[i] = network.StockPrice{
			StockID:   stockID,
			StockName: stockIDToName[stockID],
			Price:     price,
		}
		i++
	}
	//sort by stock name in lexicographically decreasing order
	sort.SliceStable(stockPrices, func(i, j int) bool {
		return stockPrices[i].StockName > stockPrices[j].StockName
	})

	return &stockPrices, nil
}

func SendToOrderExection(buyOrder order.StockOrderInterface, sellOrder order.StockOrderInterface) (network.ExecutorToMatchingEngineJSON, error) {
	buyQty := buyOrder.GetQuantity()
	sellQty := sellOrder.GetQuantity()
	quantity := buyQty
	if sellQty < buyQty {
		quantity = sellQty
	}
	transferEntity := network.MatchingEngineToExecutionJSON{
		BuyerID:       buyOrder.GetUserID(),
		SellerID:      sellOrder.GetUserID(),
		StockID:       buyOrder.GetStockID(),
		BuyOrderID:    buyOrder.GetId(),
		SellOrderID:   sellOrder.GetId(),
		IsBuyPartial:  buyQty > sellQty,
		IsSellPartial: buyQty < sellQty,
		StockPrice:    sellOrder.GetPrice(),
		Quantity:      quantity,
	}

	data, err := _networkHttpManager.OrderExecutor().Post("executor", transferEntity)

	if err != nil {
		println("Error: ", err.Error())
		return network.ExecutorToMatchingEngineJSON{}, err
	}
	print("Matched Data: ", string(data))
	var matchedData network.ExecutorToMatchingEngineJSON
	// matchedData = network.ExecutorToMatchingEngineJSON{
	// 	IsBuyFailure:  false,
	// 	IsSellFailure: false,
	// }
	err = json.Unmarshal(data, &matchedData)
	if err != nil {
		println("Error: ", err.Error())
		return network.ExecutorToMatchingEngineJSON{}, err
	}
	return matchedData, nil
}
