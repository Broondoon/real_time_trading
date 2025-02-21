package matchingEngine

import (
	"MatchingEngineService/matchingEngineStructures"
	"Shared/entities/order"
	"databaseAccessStockOrder"
	"fmt"
)

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
	orderChannel        chan order.StockOrderInterface
	updateChannel       chan *UpdateParams
	SendToOrderExection func(buyOrder order.StockOrderInterface, sellOrder order.StockOrderInterface) string
	//dirty fix
	DatabaseManager databaseAccessStockOrder.DatabaseAccessInterface
}

type NewMatchingEngineParams struct {
	StockID                  string
	InitalOrders             *[]order.StockOrderInterface
	SendToOrderExecutionFunc func(buyOrder order.StockOrderInterface, sellOrder order.StockOrderInterface) string
	DatabaseManager          databaseAccessStockOrder.DatabaseAccessInterface
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
		orderChannel:        make(chan order.StockOrderInterface, 1),
		updateChannel:       make(chan *UpdateParams, 1),
		SendToOrderExection: params.SendToOrderExecutionFunc,
		DatabaseManager:     params.DatabaseManager,
	}
	return me
}

func (me *MatchingEngine) RunMatchingEngineOrders() {
	for {

		stockOrder := <-me.orderChannel
		fmt.Println("Order received")
		if stockOrder.GetOrderType() == order.OrderTypeMarket {
			me.BuyOrderBook.AddOrder(stockOrder)
		} else {
			me.SellOrderBook.AddOrder(stockOrder)
		}
		//dequeue the top of the buy order book and sell order book
		buyOrder := me.BuyOrderBook.GetBestOrder()
		if buyOrder != nil {
			sellOrder := me.SellOrderBook.GetBestOrder()
			if sellOrder == nil {
				me.BuyOrderBook.AddOrder(buyOrder)
			} else {
				buyOrderQuantity := buyOrder.GetQuantity()
				//For now, this is gonna be a bit of a black box. The user won't know if a transaction was successful or not unless they check their transsactions.
				//later, we need to setup a rabbitMQ queue or something to send out notifications.
				for buyOrderQuantity > 0 {
					if buyOrderQuantity == sellOrder.GetQuantity() {
						//create a transaction
						result := me.SendToOrderExection(buyOrder, sellOrder)
						if result == "ERROR" {
							me.SellOrderBook.AddOrder(sellOrder)
							buyOrderQuantity = 0
						}
						switch result {
						case "COMPLETED":
							buyOrderQuantity = 0
						default:
							me.SellOrderBook.AddOrder(sellOrder)
							buyOrderQuantity = 0
						}
					} else if buyOrderQuantity < sellOrder.GetQuantity() {
						childOrder := sellOrder.CreateChildOrder(sellOrder, buyOrder)
						result := me.SendToOrderExection(buyOrder, childOrder)
						if result == "ERROR" {
							me.SellOrderBook.AddOrder(sellOrder)
							buyOrderQuantity = 0
						}
						switch result {
						case "COMPLETED":
							sellOrder.SetQuantity(sellOrder.GetQuantity() - buyOrderQuantity)
							_databaseManager.Update(sellOrder)
							buyOrderQuantity = 0
						default:
							me.SellOrderBook.AddOrder(sellOrder)
							buyOrderQuantity = 0
						}
					} else {
						childOrder := buyOrder.CreateChildOrder(buyOrder, sellOrder)
						result := me.SendToOrderExection(childOrder, sellOrder)
						if result == "ERROR" {
							me.SellOrderBook.AddOrder(sellOrder)
							buyOrderQuantity = 0
						}
						switch result {
						case "COMPLETED":
							buyOrder.SetQuantity(buyOrder.GetQuantity() - sellOrder.GetQuantity())
							buyOrderQuantity -= sellOrder.GetQuantity()
							_databaseManager.Delete(sellOrder.GetId())
							sellOrder = me.SellOrderBook.GetBestOrder()
							if sellOrder == nil {
								me.BuyOrderBook.AddOrder(buyOrder)
								break
							}
						default:
							me.SellOrderBook.AddOrder(sellOrder)
							buyOrderQuantity = 0
						}
					}
				}
				if buyOrderQuantity <= 0 {
					println("finishing Order: ", buyOrder.GetId())
					_databaseManager.Delete(buyOrder.GetId())
				}
			}
			me.SellOrderBook.CompleteBestOrderExtraction()
		}
		me.BuyOrderBook.CompleteBestOrderExtraction()
	}
}

type UpdateParams struct {
	OrderID  string
	PriceKey float64
}

func (me *MatchingEngine) RunMatchingEngineUpdates() {
	for {
		updateParams := <-me.updateChannel
		fmt.Println("Removing Order")
		me.SellOrderBook.RemoveOrder(&matchingEngineStructures.RemoveParams{
			OrderID:  updateParams.OrderID,
			PriceKey: updateParams.PriceKey,
		})
	}
}

func (me *MatchingEngine) AddOrder(order order.StockOrderInterface) {
	me.orderChannel <- order
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
