package matchingEngineStructures

import (
	"Shared/entities/order"
	"container/list"

	"github.com/google/uuid"
)

// Base Structure
type BaseOrderBookDataStructureInterface interface {
}
type BaseOrderBookDataStructure struct {
}
type NewOrderBookDataStructureParams struct {
}

func NewBaseOrderBookDataStructure(params *NewOrderBookDataStructureParams) BaseOrderBookDataStructureInterface {
	return &BaseOrderBookDataStructure{}
}

type OrderBookDataStructureInterface interface {
	BaseOrderBookDataStructureInterface
	Push(stockOrder order.StockOrderInterface)
	PushFront(stockOrder order.StockOrderInterface)
	PopNext() order.StockOrderInterface
	Remove(params *RemoveParams) order.StockOrderInterface
	Length() int
	GetPrice() float64
}

type RemoveParams struct {
	OrderID  *uuid.UUID
	PriceKey float64
}

// Queue Structure, where adding is immediatly added to the back. Removing is done from the front.

type Queue struct {
	BaseOrderBookDataStructureInterface
	data *list.List
}

func (q *Queue) Push(stockOrder order.StockOrderInterface) {
	q.data.PushBack(stockOrder)
}

func (q *Queue) PushFront(stockOrder order.StockOrderInterface) {
	q.data.PushFront(stockOrder)
}

func (q *Queue) PopNext() order.StockOrderInterface {
	e := q.data.Front()
	if e != nil {
		q.data.Remove(e)
		return e.Value.(order.StockOrderInterface)
	}
	return nil
}

func (q *Queue) Remove(params *RemoveParams) order.StockOrderInterface {
	for e := q.data.Front(); e != nil; e = e.Next() {
		if e.Value.(order.StockOrderInterface).GetIdString() == params.OrderID.String() {
			q.data.Remove(e)
			return e.Value.(order.StockOrderInterface)
		}
	}
	return nil
}

func (q *Queue) Length() int {
	return q.data.Len()
}

type NewQueueParams struct {
	*NewOrderBookDataStructureParams
}

func NewQueue(params *NewQueueParams) OrderBookDataStructureInterface {
	return &Queue{
		BaseOrderBookDataStructureInterface: NewBaseOrderBookDataStructure(params.NewOrderBookDataStructureParams),
		data:                                list.New(),
	}
}

func (q *Queue) GetPrice() float64 {
	if q.data.Len() == 0 {
		return 0
	}
	return q.data.Front().Value.(order.StockOrderInterface).GetPrice()
}

// PriceNodeMap Structure, where adding is added according to a logic system, and removing is based on a key system.
// In truth, this is more going to resemble a hash map, with the key being the price.
type PriceNodeMap struct {
	BaseOrderBookDataStructureInterface
	data             map[float64]*PriceNode
	currentBestPrice float64
}

func (p *PriceNodeMap) ensureNodeExists(key float64) {
	if _, ok := p.data[key]; !ok {
		p.data[key] = &PriceNode{
			priceList: NewQueue(&NewQueueParams{
				NewOrderBookDataStructureParams: &NewOrderBookDataStructureParams{},
			}),
			priceValue: key,
		}
		if key < p.currentBestPrice || p.currentBestPrice == 0 {
			p.currentBestPrice = key
		}
	}
}

func (p *PriceNodeMap) validateNode(node *PriceNode) {
	if node.priceList.Length() == 0 {
		p.currentBestPrice = 0
		for key := range p.data {
			if key < p.currentBestPrice || p.currentBestPrice == 0 {
				p.currentBestPrice = key
			}
		}
		delete(p.data, node.priceValue)
	}
}

func (p *PriceNodeMap) Push(stockOrder order.StockOrderInterface) {
	price := stockOrder.GetPrice()
	p.ensureNodeExists(price)
	p.data[price].priceList.Push(stockOrder)

}

func (p *PriceNodeMap) PushFront(stockOrder order.StockOrderInterface) {
	price := float64(stockOrder.GetPrice())
	p.ensureNodeExists(price)
	p.data[price].priceList.PushFront(stockOrder)
}

func (p *PriceNodeMap) PopNext() order.StockOrderInterface {
	if node, ok := p.data[p.currentBestPrice]; ok {
		order := node.priceList.PopNext()
		p.validateNode(node)
		return order
	}
	return nil
}

func (p *PriceNodeMap) Remove(params *RemoveParams) order.StockOrderInterface {
	price := params.PriceKey
	if node, ok := p.data[price]; ok {
		order := node.priceList.Remove(params)
		p.validateNode(node)
		return order
	}
	return nil
}

func (p *PriceNodeMap) Length() int {
	return len(p.data)
}

type NewPriceNodeMapParams struct {
	*NewOrderBookDataStructureParams
}

func NewPriceNodeMap(params *NewPriceNodeMapParams) OrderBookDataStructureInterface {
	return &PriceNodeMap{
		BaseOrderBookDataStructureInterface: NewBaseOrderBookDataStructure(params.NewOrderBookDataStructureParams),
		data:                                make(map[float64]*PriceNode),
		currentBestPrice:                    0,
	}
}

func (p *PriceNodeMap) GetPrice() float64 {
	return p.currentBestPrice
}

// PriceNode Structure, for internal usage on the PriceNodeMap.
type PriceNode struct {
	priceList  OrderBookDataStructureInterface
	priceValue float64
}
