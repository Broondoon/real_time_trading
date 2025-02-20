package matchingEngine

import (
	"Shared/entities/order"
	"Shared/entities/transaction"
	"Shared/network"
	"databaseAccessStock"
	"databaseAccessStockOrder"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"

	"gorm.io/gorm"
)

var _matchingEngineMap map[string]MatchingEngineInterface
var _databaseManager databaseAccessStockOrder.DatabaseAccessInterface
var _networkManager network.NetworkInterface
var _stockDatabaseAccess databaseAccessStock.DatabaseAccessInterface

func InitalizeHandlers(stockIDs *[]string,
	networkManager network.NetworkInterface, databaseManager databaseAccessStockOrder.DatabaseAccessInterface, stockDatabaseAccess databaseAccessStock.DatabaseAccessInterface) {
	_databaseManager = databaseManager
	_networkManager = networkManager
	_stockDatabaseAccess = stockDatabaseAccess
	_matchingEngineMap = make(map[string]MatchingEngineInterface)
	//Create all matching engines for stocks.
	for _, stockID := range *stockIDs {
		AddNewStock(stockID)
	}

	//Add handlers
	networkManager.AddHandleFuncUnprotected(network.HandlerParams{Pattern: "createStock", Handler: AddNewStockHandler})
	networkManager.AddHandleFuncUnprotected(network.HandlerParams{Pattern: "placeStockOrder", Handler: PlaceStockOrderHandler})
	networkManager.AddHandleFuncUnprotected(network.HandlerParams{Pattern: "deleteOrder/", Handler: DeleteStockOrderHandler})
	networkManager.AddHandleFuncProtected(network.HandlerParams{Pattern: os.Getenv("transaction_route") + "/getStockPrices", Handler: GetStockPricesHandler})
	http.HandleFunc("/health", healthHandler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Simple check: you might expand this to test database connectivity, etc.
	w.WriteHeader(http.StatusOK)
	fmt.Println(w, "OK")
}

// Expected input is a stock ID in the body of the request
// we're expecting {"StockID":"{id value}"}
func AddNewStockHandler(responseWriter http.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	var stockID network.StockID
	err := json.Unmarshal(data, &stockID)
	if err != nil {
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}

	AddNewStock(stockID.StockID)
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

func PlaceStockOrderHandler(responseWriter http.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
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

	_databaseManager.Create(stockOrder)
	me.AddOrder(stockOrder)
	return true
}

func DeleteStockOrderHandler(responseWriter http.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	orderID := queryParams.Get("id")
	err := DeleteStockOrder(orderID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		responseWriter.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	responseWriter.WriteHeader(http.StatusOK)
}

func DeleteStockOrder(orderID string) error {
	order, err := _databaseManager.GetByID(orderID)
	if err != nil {
		return err
	}
	err = _databaseManager.Delete(orderID)
	if err != nil {
		return err
	}
	me, ok := _matchingEngineMap[order.GetStockID()]
	if !ok {
		return nil
	}
	me.RemoveOrder(orderID, order.GetPrice())
	return nil
}

func GetStockPricesHandler(responseWriter http.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	prices, err := GetStockPrices()
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	pricesJSON, err := json.Marshal(prices)
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	responseWriter.Write(pricesJSON)

}

func GetStockPrices() (*[]network.StockPrice, error) {
	stocks, err := _stockDatabaseAccess.GetAll()
	if err != nil {
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

func SendToOrderExection(buyOrder order.StockOrderInterface, sellOrder order.StockOrderInterface) transaction.StockTransactionInterface {
	buyQty := buyOrder.GetQuantity()
	sellQty := sellOrder.GetQuantity()
	quantity := buyQty
	if sellQty < buyQty {
		quantity = sellQty
	}
	transferEntity := network.MatchingEngineToExecutionJSON{
		StockID:       buyOrder.GetStockID(),
		BuyOrderID:    buyOrder.GetId(),
		SellOrderID:   sellOrder.GetId(),
		IsBuyPartial:  buyQty > sellQty,
		IsSellPartial: buyQty < sellQty,
		StockPrice:    sellOrder.GetPrice(),
		Quantity:      quantity,
	}

	//need to figure out how to get the user IDs from the orders

	data, err := _networkManager.OrderExecutor().Post("orderexecutor", transferEntity)
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
