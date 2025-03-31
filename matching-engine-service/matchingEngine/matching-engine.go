package matchingEngine

import (
	"MatchingEngineService/matchingEngineStructures"
	"Shared/entities/order"
	"Shared/network"
	subfunctions "Shared/subfunctions/Multithreading"
	"databaseAccessStockOrder"
	"log"
	"sync"
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
	GetPrice() network.StockPrice
	CompletePairedOrder(params network.ExecutorToMatchingEngineJSON, buyOrder order.StockOrderInterface, sellOrder order.StockOrderInterface, quantity int, err error)
}

type MatchingEngine struct {
	StockId             string
	BuyOrderBook        matchingEngineStructures.BuyOrderBookInterface
	SellOrderBook       matchingEngineStructures.SellOrderBookInterface
	orderChannel        chan int
	updateChannel       chan *UpdateParams
	SendToOrderExection func(buyOrder order.StockOrderInterface, sellOrder order.StockOrderInterface, quantity int)
	//dirty fix
	DatabaseManager   databaseAccessStockOrder.DatabaseAccessInterface
	StockName         string
	orderState        map[string]int
	UpdateBulkRoutine subfunctions.BulkRoutineInterface[order.StockOrderInterface]
	mutex             *sync.Mutex
}

type NewMatchingEngineParams struct {
	StockID                  string
	InitalOrders             *[]order.StockOrderInterface
	SendToOrderExecutionFunc func(buyOrder order.StockOrderInterface, sellOrder order.StockOrderInterface, quantity int)
	DatabaseManager          databaseAccessStockOrder.DatabaseAccessInterface
	StockName                string
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
		orderChannel:        make(chan int, 5000),
		updateChannel:       make(chan *UpdateParams, 1000),
		SendToOrderExection: params.SendToOrderExecutionFunc,
		DatabaseManager:     params.DatabaseManager,
		StockName:           params.StockName,
		orderState:          make(map[string]int),
		mutex:               &sync.Mutex{},
	}

	updateFunc := func(data *[]order.StockOrderInterface, TransferParams any) error {
		errorList, err := me.DatabaseManager.UpdateBulk(data)
		if err != nil {
			log.Println("Error in deleting bulk orders")
			return err
		}
		for id, err := range errorList {
			log.Println("Failed to update stock order: ", id, " with Error Code: ", err)
		}
		return nil
	}

	me.UpdateBulkRoutine = subfunctions.NewBulkRoutine(&subfunctions.BulkRoutineParams[order.StockOrderInterface]{
		Routine: updateFunc,
	})

	return me
}

func (me *MatchingEngine) RunMatchingEngineOrders() {
	var buyOrder order.StockOrderInterface
	var sellOrder order.StockOrderInterface
	var haveBuyOrder bool
	var haveSellOrder bool
	for {
		//dequeue the top of the buy order book and sell order book
		me.mutex.Lock()
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
					//log.Println("Buy Order is nil. Returning sell order")
					me.SellOrderBook.AddOrder(sellOrder)
					sellOrder = nil
				}
			} else if sellOrder == nil {
				haveBuyOrder = true
				haveSellOrder = false
				//log.Println("Sell Order is nil, Returning buy order")
				me.BuyOrderBook.ReturnOrder(buyOrder)
				buyOrder = nil
			}
		}
		me.mutex.Unlock()
		if buyOrder != nil && sellOrder != nil {
			// buyIsChild := false
			// sellIsChild := false
			// var parentOrder order.StockOrderInterface
			// if buyOrder.GetQuantity() < sellOrder.GetQuantity() {
			// 	parentOrder = sellOrder
			// 	sellIsChild = true
			// 	sellOrder = sellOrder.CreateChildOrder(sellOrder, buyOrder)
			// 	if sellOrder.GetQuantity() == parentOrder.GetQuantity() {
			// 		log.Println("Sell Order is nil, Returning buy order")
			// 		close(me.orderChannel)
			// 		close(me.updateChannel)
			// 	}
			// }
			// if buyOrder.GetQuantity() > sellOrder.GetQuantity() {
			// 	parentOrder = buyOrder
			// 	buyIsChild = true
			// 	buyOrder = buyOrder.CreateChildOrder(buyOrder, sellOrder)
			// 	if sellOrder.GetQuantity() == parentOrder.GetQuantity() {
			// 	}
			// }
			me.mutex.Lock()
			if _, ok := me.orderState[sellOrder.GetIdString()]; ok {
				me.orderState[sellOrder.GetIdString()]++
			} else {
				me.orderState[sellOrder.GetIdString()] = 1
			}
			if _, ok := me.orderState[buyOrder.GetIdString()]; ok {
				me.orderState[buyOrder.GetIdString()]++
			} else {
				me.orderState[buyOrder.GetIdString()] = 1
			}
			me.mutex.Unlock()

			quantity := min(sellOrder.GetQuantity(), buyOrder.GetQuantity())
			sellOrder.UpdateQuantity(-quantity)
			buyOrder.UpdateQuantity(-quantity)
			go me.SendToOrderExection(buyOrder, sellOrder, quantity)
			if buyOrder.GetQuantity() == 0 {
				buyOrder = nil
			} else {
				sellOrder = nil
			}
		} else {
			for {
				// log.Println("No orders to match")
				// log.Println("Waiting for order")
				orderReceived := <-me.orderChannel
				// log.Println("Order received")
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

func (me *MatchingEngine) GetPrice() network.StockPrice {
	return network.StockPrice{
		StockID:   me.StockId,
		Price:     me.SellOrderBook.GetBestPrice(),
		StockName: me.StockName,
	}

}

func (me *MatchingEngine) CompletePairedOrder(params network.ExecutorToMatchingEngineJSON, buyOrder order.StockOrderInterface, sellOrder order.StockOrderInterface, quantity int, err error) {
	if err != nil {
		log.Println("Error in order execution: ", err)
	}
	if params.IsBuyFailure || params.IsSellFailure {
		me.mutex.Lock()
		if params.IsBuyFailure {
			sellOrder.UpdateQuantity(quantity)
			me.orderState[sellOrder.GetIdString()]--
			if me.orderState[sellOrder.GetIdString()] == 0 {
				me.SellOrderBook.ReturnOrder(sellOrder)
				me.orderChannel <- ISSELL
			}
		}
		if params.IsBuyFailure {
			buyOrder.UpdateQuantity(quantity)
			me.orderState[buyOrder.GetIdString()]--
			if me.orderState[buyOrder.GetIdString()] == 0 {
				me.BuyOrderBook.ReturnOrder(buyOrder)
				me.orderChannel <- ISBUY
			}
		}
		me.mutex.Unlock()
	} else {
		sellOrder.UpdateDBQuantity(-quantity)
		me.UpdateBulkRoutine.Insert(sellOrder)
		buyOrder.UpdateDBQuantity(-quantity)
		me.UpdateBulkRoutine.Insert(buyOrder)
	}
}
