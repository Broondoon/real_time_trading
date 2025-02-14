package matchingEngine

import (
	"Shared/entities/order"
	"Shared/entities/stock"
	"Shared/network"
	"net/http"
)

var _matchingEngineMap map[string]MatchingEngineInterface

func InitalizeHandlers(stockList []stock.StockInterface,
	networkManager network.HttpClientInterface) {
	//Create all matching engines for stocks.
	for _, stock := range stockList {
		AddNewStock(stock)
		RepopulateStocksFromDatabase(stock)
	}

	networkManager.AddHandleFunc(network.HandlerParams{Pattern: "/createStock", Handler: AddNewStockHandler})

	networkManager.Listen(network.ListenerParams{
		Port:    "8080",
		Handler: nil,
	})
}

func AddNewStockHandler(responseWriter http.ResponseWriter, data []byte) {
	s, err := stock.Parse(data)
	if err != nil {
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	AddNewStock(s)
}

func AddNewStock(stock stock.StockInterface) {
	_, ok := _matchingEngineMap[stock.GetId()]
	//if we don't have a matching engine for this stock, create one
	if !ok {
		me := NewMatchingEngineForStock(NewMatchingEngineParams{Stock: stock})
		_matchingEngineMap[stock.GetId()] = me
		go me.RunMatchingEngineOrders()
		go me.RunMatchingEngineUpdates()
	}
}

func RepopulateStocksFromDatabase(stock stock.StockInterface) {

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
	me.AddOrder(stockOrder)
	return true
}
