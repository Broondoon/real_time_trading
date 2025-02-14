package matchingEngineStructures

import (
	"Shared/entities/order"
	"container/list"
	"errors"
)

type OrderBookDataStructureInterface interface {
	Push(stockOrder order.StockOrderInterface)
	PushFront(stockOrder order.StockOrderInterface)
	PopNext() order.StockOrderInterface
	Remove(params RemoveParams) order.StockOrderInterface
	Length() int
}

// Base Structure
type OrderBookDataStructure struct {
}

func (o *OrderBookDataStructure) Push(stockOrder order.StockOrderInterface) {
	panic(errors.New("uninitialized method: Push"))
}

func (o *OrderBookDataStructure) PushFront(stockOrder order.StockOrderInterface) {
	panic(errors.New("uninitialized method: PushFront"))
}

func (o *OrderBookDataStructure) PopNext() order.StockOrderInterface {
	panic(errors.New("uninitialized method: PopNext"))
}

type RemoveParams struct {
	StockId  string
	priceKey float64
}

func (o *OrderBookDataStructure) Remove(params RemoveParams) order.StockOrderInterface {
	panic(errors.New("uninitialized method: Remove"))
}

func (o *OrderBookDataStructure) Length() int {
	panic(errors.New("uninitialized method: Length"))
}

type NewOrderBookDataStructureParams struct {
}

func NewOrderBookDataStructure(params NewOrderBookDataStructureParams) *OrderBookDataStructure {
	return &OrderBookDataStructure{}
}

// Queue Structure, where adding is immediatly added to the back. Removing is done from the front.
type Queue struct {
	OrderBookDataStructureInterface
	data list.List
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
	}
	return e.Value.(order.StockOrderInterface)
}

func (q *Queue) Remove(params RemoveParams) order.StockOrderInterface {
	for e := q.data.Front(); e != nil; e = e.Next() {
		if e.Value.(order.StockOrderInterface).GetId() == params.StockId {
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
	NewOrderBookDataStructureParams
}

func NewQueue(params NewQueueParams) *Queue {
	return &Queue{
		OrderBookDataStructureInterface: NewOrderBookDataStructure(params.NewOrderBookDataStructureParams),
	}
}

// PriceNodeMap Structure, where adding is added according to a logic system, and removing is based on a key system.
// In truth, this is more going to resemble a hash map, with the key being the price.
type PriceNodeMap struct {
	OrderBookDataStructureInterface
	data             map[float64]PriceNode
	currentBestPrice float64
}

func (p *PriceNodeMap) ensureNodeExists(key float64) {
	if _, ok := p.data[key]; !ok {
		p.data[key] = PriceNode{
			priceList:  NewQueue(NewQueueParams{NewOrderBookDataStructureParams{}}),
			priceValue: key,
		}
		if key > p.currentBestPrice {
			p.currentBestPrice = key
		}
	}
}

func (p *PriceNodeMap) validateNode(node PriceNode) {
	if node.priceList.Length() == 0 {
		delete(p.data, p.currentBestPrice)
		p.currentBestPrice = 0
		for key := range p.data {
			if key > p.currentBestPrice {
				p.currentBestPrice = key
			}
		}
	}
}

func (p *PriceNodeMap) Push(stockOrder order.StockOrderInterface) {
	price := float64(stockOrder.GetPrice())
	p.ensureNodeExists(price)
	p.data[price].priceList.Push(stockOrder)

}

func (p *PriceNodeMap) PushFront(stockOrder order.StockOrderInterface) {
	price := float64(stockOrder.GetPrice())
	p.ensureNodeExists(price)
	p.data[price].priceList.PushFront(stockOrder)
}

func (p *PriceNodeMap) PopNext() order.StockOrderInterface {
	node := p.data[p.currentBestPrice]
	order := node.priceList.PopNext()
	p.validateNode(node)
	return order
}

func (p *PriceNodeMap) Remove(params RemoveParams) order.StockOrderInterface {
	price := params.priceKey
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
	NewOrderBookDataStructureParams
}

func NewPriceNodeMap(params NewPriceNodeMapParams) *PriceNodeMap {
	return &PriceNodeMap{
		OrderBookDataStructureInterface: NewOrderBookDataStructure(params.NewOrderBookDataStructureParams),
		data:                            make(map[float64]PriceNode),
	}
}

// PriceNode Structure, for internal usage on the PriceNodeMap.
type PriceNode struct {
	priceList  OrderBookDataStructureInterface
	priceValue float64
}
