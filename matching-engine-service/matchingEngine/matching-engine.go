package matchingEngine

import (
	"MatchingEngineService/matchingEngineStructures"
	"Shared/entities/order"
	"Shared/network"
	"databaseAccessStockOrder"
	"log"
)

const ISBUY = 0
const ISSELL = 1

// https://gobyexample.com/channels
// https://chatgpt.com/share/67aa804e-4678-8006-970a-23d76d933f3c
type MatchingEngineInterface interface {
	AddOrder(stockOrder order.StockOrderInterface)
	RemoveOrder(orderID string, priceKey float64)
	RunMatchingEngineOrders()
	RunMatchingEngineUpdates()
	GetPrice() float64
}

type MatchingEngine struct {
	StockId             string
	BuyOrderBook        matchingEngineStructures.BuyOrderBookInterface
	SellOrderBook       matchingEngineStructures.SellOrderBookInterface
	orderChannel        chan int
	updateChannel       chan *UpdateParams
	SendToOrderExection func(buyOrder order.StockOrderInterface, sellOrder order.StockOrderInterface) (network.ExecutorToMatchingEngineJSON, error)
	//dirty fix
	DatabaseManager databaseAccessStockOrder.DatabaseAccessInterface
	UpdatePrice     chan network.StockPrice
}

type NewMatchingEngineParams struct {
	StockID                  string
	InitalOrders             *[]order.StockOrderInterface
	SendToOrderExecutionFunc func(buyOrder order.StockOrderInterface, sellOrder order.StockOrderInterface) (network.ExecutorToMatchingEngineJSON, error)
	DatabaseManager          databaseAccessStockOrder.DatabaseAccessInterface
	UpdatePrice              chan network.StockPrice
}

func NewMatchingEngineForStock(params *NewMatchingEngineParams) MatchingEngineInterface {
	var marketOrders []order.StockOrderInterface
	var limitOrders []order.StockOrderInterface
	for _, order := range *params.InitalOrders {
		if order.GetIsBuy() {
			marketOrders = append(marketOrders, order)
		} else {
			limitOrders = append(limitOrders, order)
		}
	}
	me := &MatchingEngine{
		StockId:             params.StockID,
		BuyOrderBook:        matchingEngineStructures.DefaultBuyOrderBook(&marketOrders),
		SellOrderBook:       matchingEngineStructures.DefaultSellOrderBook(&limitOrders),
		orderChannel:        make(chan int, 1000),
		updateChannel:       make(chan *UpdateParams, 1000),
		SendToOrderExection: params.SendToOrderExecutionFunc,
		DatabaseManager:     params.DatabaseManager,
		UpdatePrice:         params.UpdatePrice,
	}
	me.UpdatePrice <- network.StockPrice{
		StockID: me.StockId,
		Price:   me.GetPrice(),
	}
	return me
}

func (me *MatchingEngine) RunMatchingEngineOrders() {
	var buyOrder order.StockOrderInterface
	var sellOrder order.StockOrderInterface
	var haveBuyOrder bool
	var haveSellOrder bool
	for {
		//dequeue the top of the buy order book and sell order book
		if buyOrder == nil {
			buyOrder = me.BuyOrderBook.GetBestOrder()
		} else {
			haveBuyOrder = true
		}
		if sellOrder == nil {
			sellOrder = me.SellOrderBook.GetBestOrder()
		} else {
			haveSellOrder = true
		}
		if buyOrder == nil || sellOrder == nil {
			if buyOrder == nil {
				haveBuyOrder = false
				if sellOrder != nil {
					haveSellOrder = true
					log.Println("Buy Order is nil. Returning sell order")
					me.SellOrderBook.AddOrder(sellOrder)
					sellOrder = nil
				}
			} else if sellOrder == nil {
				haveBuyOrder = true
				haveSellOrder = false
				log.Println("Sell Order is nil, Returning buy order")
				me.BuyOrderBook.ReturnOrder(buyOrder)
				buyOrder = nil
			}
		} else {
			me.UpdatePrice <- network.StockPrice{
				StockID: me.StockId,
				Price:   sellOrder.GetPrice(),
			}
		}
		if buyOrder != nil && sellOrder != nil {
			buyIsChild := false
			sellIsChild := false
			var parentOrder order.StockOrderInterface
			if buyOrder.GetQuantity() < sellOrder.GetQuantity() {
				parentOrder = sellOrder
				sellIsChild = true
				sellOrder = sellOrder.CreateChildOrder(sellOrder, buyOrder)
				if sellOrder.GetQuantity() == parentOrder.GetQuantity() {
					close(me.orderChannel)
					close(me.updateChannel)
				}
			}
			if buyOrder.GetQuantity() > sellOrder.GetQuantity() {
				parentOrder = buyOrder
				buyIsChild = true
				buyOrder = buyOrder.CreateChildOrder(buyOrder, sellOrder)
				if sellOrder.GetQuantity() == parentOrder.GetQuantity() {
				}
			}
			result, err := me.SendToOrderExection(buyOrder, sellOrder)
			sellOrderQuantity := sellOrder.GetQuantity()
			buyOrderQuantity := buyOrder.GetQuantity()
			if sellIsChild {
				sellOrder = parentOrder
			} else if buyIsChild {
				buyOrder = parentOrder
			}
			if err != nil {
				//rollback
				me.BuyOrderBook.ReturnOrder(buyOrder)
				me.SellOrderBook.AddOrder(sellOrder)
				close(me.orderChannel)
				close(me.updateChannel)
				panic("Error in order execution")
			} else if result.IsBuyFailure {
				buyOrder = nil
			} else if result.IsSellFailure {
				sellOrder = nil
			} else {
				sellOrder.UpdateQuantity(-buyOrderQuantity)
				buyOrder.UpdateQuantity(-sellOrderQuantity)
				if sellOrder.GetQuantity() == 0 {
					_databaseManager.Delete(sellOrder.GetId())
					sellOrder = nil
				} else {
					_databaseManager.Update(sellOrder)
				}

				if buyOrder.GetQuantity() == 0 {
					_databaseManager.Delete(buyOrder.GetId())
					buyOrder = nil
				} else {
					_databaseManager.Update(buyOrder)
				}
			}
		} else {
			for {
				log.Println("No orders to match")
				log.Println("Waiting for order")
				orderReceived := <-me.orderChannel
				log.Println("Order received")
				if orderReceived == ISBUY {
					haveBuyOrder = true
					if haveSellOrder {
						break
					}
				} else if orderReceived == ISSELL {
					haveSellOrder = true
					if haveBuyOrder {
						break
					}
				}
			}
		}
	}
}

type UpdateParams struct {
	OrderID  string
	PriceKey float64
}

func (me *MatchingEngine) RunMatchingEngineUpdates() {
	for {
		updateParams := <-me.updateChannel
		me.SellOrderBook.RemoveOrder(&matchingEngineStructures.RemoveParams{
			OrderID:  updateParams.OrderID,
			PriceKey: updateParams.PriceKey,
		})
	}
}

func (me *MatchingEngine) AddOrder(stockOrder order.StockOrderInterface) {
	if stockOrder.GetOrderType() == order.OrderTypeMarket {
		me.BuyOrderBook.AddOrder(stockOrder)
		me.orderChannel <- ISBUY
	} else {
		me.SellOrderBook.AddOrder(stockOrder)
		me.orderChannel <- ISSELL
	}
}

func (me *MatchingEngine) RemoveOrder(orderID string, priceKey float64) {
	me.updateChannel <- &UpdateParams{
		OrderID:  orderID,
		PriceKey: priceKey,
	}
}

func (me *MatchingEngine) GetPrice() float64 {
	return me.SellOrderBook.GetBestPrice()
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
