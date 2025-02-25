package matchingEngine

import (
	"MatchingEngineService/matchingEngineStructures"
	"Shared/entities/order"
	"Shared/network"
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
	SendToOrderExection func(buyOrder order.StockOrderInterface, sellOrder order.StockOrderInterface) (network.ExecutorToMatchingEngineJSON, error)
	//dirty fix
	DatabaseManager databaseAccessStockOrder.DatabaseAccessInterface
}

type NewMatchingEngineParams struct {
	StockID                  string
	InitalOrders             *[]order.StockOrderInterface
	SendToOrderExecutionFunc func(buyOrder order.StockOrderInterface, sellOrder order.StockOrderInterface) (network.ExecutorToMatchingEngineJSON, error)
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
		updateChannel:       make(chan *UpdateParams),
		SendToOrderExection: params.SendToOrderExecutionFunc,
		DatabaseManager:     params.DatabaseManager,
	}
	return me
}

func (me *MatchingEngine) RunMatchingEngineOrders() {
	println("Running Matching Engine Orders")
	var buyOrder order.StockOrderInterface
	var sellOrder order.StockOrderInterface
	for {
		//dequeue the top of the buy order book and sell order book
		if buyOrder == nil {
			println("Getting best buy order")
			buyOrder = me.BuyOrderBook.GetBestOrder()
			if buyOrder != nil {
				temp, err := buyOrder.ToJSON()
				if err != nil {
					println("Error: ", err.Error())
				}
				print("Buy Order: ", string(temp))
			}
		}
		if sellOrder == nil {
			println("Getting best sell order")
			sellOrder = me.SellOrderBook.GetBestOrder()
			if sellOrder != nil {
				temp, err := sellOrder.ToJSON()
				if err != nil {
					println("Error: ", err.Error())
				}
				print("Sell Order: ", string(temp))
			}
		}
		if buyOrder == nil || sellOrder == nil {
			if buyOrder == nil {
				println("Buy Order is nil")
				if sellOrder != nil {
					println("Returning sell order")
					me.SellOrderBook.AddOrder(sellOrder)
					sellOrder = nil
				}
			} else if sellOrder == nil {
				println("Sell Order is nil")
				println("Returning buy order")
				me.BuyOrderBook.ReturnOrder(buyOrder)
				buyOrder = nil
			}
		}
		println("Starting Match")
		if buyOrder != nil && sellOrder != nil {
			println("Matching Orders. Buy Order: ", buyOrder.GetId(), " Sell Order: ", sellOrder.GetId())
			buyIsChild := false
			sellIsChild := false
			var parentOrder order.StockOrderInterface
			if buyOrder.GetQuantity() < sellOrder.GetQuantity() {
				println("Creating sell child order for: ", sellOrder.GetId())
				parentOrder = sellOrder
				sellIsChild = true
				sellOrder = sellOrder.CreateChildOrder(sellOrder, buyOrder)
				println("Parent Order Quantity: ", parentOrder.GetQuantity())
				println("Child Order Quantity: ", sellOrder.GetQuantity())
				if sellOrder.GetQuantity() == parentOrder.GetQuantity() {
					close(me.orderChannel)
					close(me.updateChannel)
					panic("Child order quantity is equal to parent order quantity. This should not happen")
				}
			}
			if buyOrder.GetQuantity() > sellOrder.GetQuantity() {
				println("Creating buy child order for: ", buyOrder.GetId())
				parentOrder = buyOrder
				buyIsChild = true
				buyOrder = buyOrder.CreateChildOrder(buyOrder, sellOrder)
				println("Parent Order Quantity: ", parentOrder.GetQuantity())
				println("Child Order Quantity: ", buyOrder.GetQuantity())
				if sellOrder.GetQuantity() == parentOrder.GetQuantity() {
					panic("Child order quantity is equal to parent order quantity. This should not happen")
				}
			}
			result, err := me.SendToOrderExection(buyOrder, sellOrder)
			println("Order Executed: ")
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
				println("Buy Order Failed: ", buyOrder.GetId())
				buyOrder = nil
			} else if result.IsSellFailure {
				println("Sell Order Failed: ", sellOrder.GetId())
				sellOrder = nil
			} else {
				println("Cleaning up orders")
				sellOrder.SetQuantity(sellOrder.GetQuantity() - buyOrderQuantity)
				buyOrder.SetQuantity(buyOrder.GetQuantity() - sellOrderQuantity)
				if sellOrder.GetQuantity() == 0 {
					println("finishing sell Order: ", buyOrder.GetId())
					_databaseManager.Delete(sellOrder.GetId())
					sellOrder = nil
				} else {
					_databaseManager.Update(sellOrder)
				}

				if buyOrder.GetQuantity() == 0 {
					println("finishing buy Order: ", buyOrder.GetId())
					_databaseManager.Delete(buyOrder.GetId())
					buyOrder = nil
				} else {
					_databaseManager.Update(buyOrder)
				}
			}
		} else {
			fmt.Println("No orders to match")
			fmt.Println("Waiting for order")
			stockOrder := <-me.orderChannel
			fmt.Println("Order received")
			temp, err := stockOrder.ToJSON()
			if err != nil {
				fmt.Println("Error: ", err.Error())
				continue
			}
			println("Order: ", string(temp))
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
		fmt.Println("Removing Order")
		me.SellOrderBook.RemoveOrder(&matchingEngineStructures.RemoveParams{
			OrderID:  updateParams.OrderID,
			PriceKey: updateParams.PriceKey,
		})
	}
}

func (me *MatchingEngine) AddOrder(stockOrder order.StockOrderInterface) {
	println("Adding Order")
	if stockOrder.GetOrderType() == order.OrderTypeMarket {
		me.BuyOrderBook.AddOrder(stockOrder)
	} else {
		me.SellOrderBook.AddOrder(stockOrder)
	}
	if len(me.orderChannel) == 0 {
		me.orderChannel <- stockOrder
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
