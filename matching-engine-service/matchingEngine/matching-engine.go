package matchingEngine

import (
	"MatchingEngineService/matchingEngineStructures"
	"Shared/entities/order"
	"Shared/entities/stock"
)

// https://gobyexample.com/channels
// https://chatgpt.com/share/67aa804e-4678-8006-970a-23d76d933f3c
type MatchingEngineInterface interface {
	AddOrder(order order.StockOrderInterface)
	RunMatchingEngineOrders()
	RunMatchingEngineUpdates()
}

type MatchingEngine struct {
	StockId       string
	BuyOrderBook  matchingEngineStructures.BuyOrderBookInterface
	SellOrderBook matchingEngineStructures.SellOrderBookInterface
	orderChannel  chan order.StockOrderInterface
	updateChannel chan UpdateMatchingEngineChannelParams
}
type UpdateMatchingEngineChannelParams struct {
}

type NewMatchingEngineParams struct {
	Stock stock.StockInterface
}

func NewMatchingEngineForStock(params NewMatchingEngineParams) *MatchingEngine {
	me := &MatchingEngine{
		StockId:       params.Stock.GetId(),
		BuyOrderBook:  matchingEngineStructures.DefaultBuyOrderBook(),
		SellOrderBook: matchingEngineStructures.DefaultSellOrderBook(),
		orderChannel:  make(chan order.StockOrderInterface, 1),
		updateChannel: make(chan UpdateMatchingEngineChannelParams, 1),
	}
	return me
}

func (me *MatchingEngine) RunMatchingEngineOrders() {
	for {

		stockOrder := <-me.orderChannel
		println("order received")
		if stockOrder.GetOrderType() == order.OrderTypeMarket {
			me.BuyOrderBook.AddOrder(stockOrder)
		} else {
			me.SellOrderBook.AddOrder(stockOrder)
		}
	}
}

func (me *MatchingEngine) RunMatchingEngineUpdates() {
	for {
		<-me.updateChannel
	}
}

func (me *MatchingEngine) AddOrder(order order.StockOrderInterface) {
	me.orderChannel <- order
}

//fake matching engine mock for testing

type FakeMatchingEngine struct {
	ordersCalled  bool
	updatesCalled bool
	ordersCh      chan struct{}
	updatesCh     chan struct{}
}

func (fme *FakeMatchingEngine) AddOrder(o order.StockOrderInterface) {}

func (fme *FakeMatchingEngine) RunMatchingEngineOrders() {
	fme.ordersCalled = true
	close(fme.ordersCh)
}

func (fme *FakeMatchingEngine) RunMatchingEngineUpdates() {
	fme.updatesCalled = true
	close(fme.updatesCh)
}

//hmmm, when we get partial matches and need to reinsert orders, taht might cause race conditions if we need to undo or reinsert orders...
//we might need to pair pops with locking and then later unlocking on this side.
