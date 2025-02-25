package matchingEngineStructures

//Methods partially implmeneted using Copilot and ChatGPT 03-mini (preview)
import (
	"Shared/entities/order"
	"sort"
	"sync"
)

type OrderBookInterface interface {
	GetBestOrder() order.StockOrderInterface
	GetNextOrder() order.StockOrderInterface
	CompleteBestOrderExtraction()
	AddOrder(stockOrder order.StockOrderInterface)
	ReturnOrder(stockOrder order.StockOrderInterface)
	GetMutex() *sync.Mutex
	GetData() OrderBookDataStructureInterface
	GetBestPrice() float64
}

type OrderBook struct {
	data  OrderBookDataStructureInterface
	mutex *sync.Mutex
}

func (o *OrderBook) GetData() OrderBookDataStructureInterface {
	return o.data
}

func (o *OrderBook) GetMutex() *sync.Mutex {
	return o.mutex
}

func (o *OrderBook) GetBestPrice() float64 {
	return o.GetData().GetPrice()
}

// potential race condition here. if we need to actually put the order back due to complications, while it was extracted, other orders could have been extracted.
// so current half solution is to only unlock after the order is extracted and we are sure we are done with it.
func (o *OrderBook) GetBestOrder() order.StockOrderInterface {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	return o.data.PopNext()
}
func (o *OrderBook) GetNextOrder() order.StockOrderInterface {
	return o.data.PopNext()
}

func (o *OrderBook) CompleteBestOrderExtraction() {
	o.mutex.Unlock()
}

func (o *OrderBook) AddOrder(stockOrder order.StockOrderInterface) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.data.Push(stockOrder)
}

func (o *OrderBook) ReturnOrder(stockOrder order.StockOrderInterface) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.data.PushFront(stockOrder)
}

type NewOrderBookParams struct {
	dataStructure OrderBookDataStructureInterface
	InitalOrders  *[]order.StockOrderInterface
}

func NewOrderBook(params *NewOrderBookParams) OrderBookInterface {
	ob := &OrderBook{
		data: params.dataStructure,
	}
	for _, order := range *params.InitalOrders {
		ob.AddOrder(order)
	}
	return ob
}

type BuyOrderBookInterface interface {
	OrderBookInterface
}

type BuyOrderBook struct {
	OrderBookInterface
}

type NewBuyOrderBookParams struct {
	*NewOrderBookParams
}

func NewBuyOrderBook(params *NewBuyOrderBookParams) BuyOrderBookInterface {
	return &BuyOrderBook{
		OrderBookInterface: NewOrderBook(params.NewOrderBookParams),
	}
}

func DefaultBuyOrderBook(initalOrders *[]order.StockOrderInterface) BuyOrderBookInterface {
	//sort initial orders by date created, Oldest to newest
	sort.SliceStable((*initalOrders), func(i, j int) bool {
		return (*initalOrders)[i].GetDateCreated().Before((*initalOrders)[j].GetDateCreated())
	})
	return NewBuyOrderBook(&NewBuyOrderBookParams{
		&NewOrderBookParams{
			dataStructure: NewQueue(&NewQueueParams{
				&NewOrderBookDataStructureParams{},
			}),
			InitalOrders: initalOrders,
		},
	})
}

type SellOrderBookInterface interface {
	OrderBookInterface
	RemoveOrder(params *RemoveParams) order.StockOrderInterface
}

type SellOrderBook struct {
	OrderBookInterface
}

func (s *SellOrderBook) RemoveOrder(params *RemoveParams) order.StockOrderInterface {
	s.GetMutex().Lock()
	defer s.GetMutex().Unlock()
	return s.GetData().Remove(params)
}

type NewSellOrderBookParams struct {
	*NewOrderBookParams // Leave empty for default
}

func NewSellOrderBook(params *NewSellOrderBookParams) SellOrderBookInterface {
	return &SellOrderBook{
		OrderBookInterface: NewOrderBook(params.NewOrderBookParams),
	}
}

func DefaultSellOrderBook(initalOrders *[]order.StockOrderInterface) SellOrderBookInterface {
	return NewSellOrderBook(&NewSellOrderBookParams{
		&NewOrderBookParams{
			dataStructure: NewPriceNodeMap(&NewPriceNodeMapParams{
				&NewOrderBookDataStructureParams{},
			}),
			InitalOrders: initalOrders,
		},
	})
}

type FakeOrderBook struct {
	BestOrder       order.StockOrderInterface
	BestOrderCalled bool
	AddOrderCalled  bool
	MutextState     bool
}

func (fob *FakeOrderBook) GetBestOrder() order.StockOrderInterface {
	fob.BestOrderCalled = true
	if !fob.MutextState {
		fob.MutextState = true
		return fob.BestOrder
	}
	return nil
}

func (fob *FakeOrderBook) CompleteBestOrderExtraction() {
	fob.MutextState = false
}

func (fob *FakeOrderBook) AddOrder(stockOrder order.StockOrderInterface) {
	fob.AddOrderCalled = true
	if !fob.MutextState {
		fob.MutextState = true
		if fob.BestOrder.GetPrice() < stockOrder.GetPrice() {
			fob.BestOrder = stockOrder
		}
		fob.MutextState = false
	}
}
