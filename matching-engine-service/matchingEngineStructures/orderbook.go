package matchingEngineStructures

//Methods partially implmeneted using Copilot and ChatGPT 03-mini (preview)
import (
	"Shared/entities/order"
	"sync"
)

type OrderBookInterface interface {
	GetBestOrder() order.StockOrderInterface
	CompleteBestOrderExtraction()
	AddOrder(stockOrder order.StockOrderInterface)
	GetMutex() *sync.Mutex
	GetData() OrderBookDataStructureInterface
}

type OrderBook struct {
	data  OrderBookDataStructureInterface
	mutex sync.Mutex
}

func (o *OrderBook) GetData() OrderBookDataStructureInterface {
	return o.data
}

func (o *OrderBook) GetMutex() *sync.Mutex {
	return &o.mutex
}

// potential race condition here. if we need to actually put the order back due to complications, while it was extracted, other orders could have been extracted.
// so current half solution is to only unlock after the order is extracted and we are sure we are done with it.
func (o *OrderBook) GetBestOrder() order.StockOrderInterface {
	o.mutex.Lock()
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

type NewOrderBookParams struct {
	dataStructure OrderBookDataStructureInterface // Leave empty for default
}

func NewOrderBook(params NewOrderBookParams) *OrderBook {
	return &OrderBook{
		data: params.dataStructure,
	}
}

type BuyOrderBookInterface interface {
	OrderBookInterface
}

type BuyOrderBook struct {
	OrderBookInterface
}

type NewBuyOrderBookParams struct {
	NewOrderBookParams // Leave empty for default
}

func NewBuyOrderBook(params NewBuyOrderBookParams) *BuyOrderBook {

	return &BuyOrderBook{
		OrderBookInterface: NewOrderBook(params.NewOrderBookParams),
	}
}

func DefaultBuyOrderBook() *BuyOrderBook {
	return NewBuyOrderBook(NewBuyOrderBookParams{NewOrderBookParams{dataStructure: NewQueue(NewQueueParams{NewOrderBookDataStructureParams{}})}})
}

type SellOrderBookInterface interface {
	OrderBookInterface
	RemoveOrder(params RemoveParams) order.StockOrderInterface
}

type SellOrderBook struct {
	OrderBookInterface
}

func (s *SellOrderBook) RemoveOrder(params RemoveParams) order.StockOrderInterface {
	s.GetMutex().Lock()
	defer s.GetMutex().Unlock()
	return s.GetData().Remove(params)
}

type NewSellOrderBookParams struct {
	NewOrderBookParams // Leave empty for default
}

func NewSellOrderBook(params NewSellOrderBookParams) *SellOrderBook {
	return &SellOrderBook{
		OrderBookInterface: NewOrderBook(params.NewOrderBookParams),
	}
}

func DefaultSellOrderBook() *SellOrderBook {
	return NewSellOrderBook(NewSellOrderBookParams{NewOrderBookParams{dataStructure: NewPriceNodeMap(NewPriceNodeMapParams{NewOrderBookDataStructureParams{}})}})
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
