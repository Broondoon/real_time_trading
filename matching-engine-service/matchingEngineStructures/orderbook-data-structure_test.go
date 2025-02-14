package matchingEngineStructures

//Created using Copilot and ChatGPT 03-mini (preview)
//TODO: FIX TESTS
import (
	"testing"

	"Shared/entities/order"
)

// --- Helper: Create fake orders using FakeStockOrder ---
func newFakeOrder(id string, price float64) order.StockOrderInterface {
	// Using FakeStockOrder from Shared/entities/order/StockOrder.go
	return &order.FakeStockOrder{
		StockID:   id,
		IsBuy:     true,
		OrderType: order.OrderTypeLimit,
		Quantity:  10,
		Price:     price,
	}
}

// --------------------- Queue Tests ---------------------
func TestQueuePushAndPop(t *testing.T) {
	// Initialize a new Queue.
	q := NewQueue(NewQueueParams{NewOrderBookDataStructureParams{}})

	// Create fake orders.
	order1 := newFakeOrder("order1", 100.0)
	order2 := newFakeOrder("order2", 100.0)

	// Test Push.
	q.Push(order1)
	q.Push(order2)
	if q.Length() != 2 {
		t.Errorf("Expected length 2, got %d", q.Length())
	}

	// Test PopNext (FIFO).
	popped := q.PopNext()
	if popped.GetId() != "order1" {
		t.Errorf("Expected first order id 'order1', got %s", popped.GetId())
	}
	if q.Length() != 1 {
		t.Errorf("Expected length 1 after pop, got %d", q.Length())
	}
}

func TestQueuePushFrontAndRemove(t *testing.T) {
	q := NewQueue(NewQueueParams{NewOrderBookDataStructureParams{}})

	order1 := newFakeOrder("order1", 100.0)
	order2 := newFakeOrder("order2", 100.0)
	order3 := newFakeOrder("order3", 100.0)

	// Push order1 then order2.
	q.Push(order1)
	q.Push(order2)
	// PushFront order3; now order3 should be at the front.
	q.PushFront(order3)
	if q.Length() != 3 {
		t.Errorf("Expected length 3, got %d", q.Length())
	}

	// Remove order2.
	removed := q.Remove(RemoveParams{StockId: "order2"})
	if removed == nil || removed.GetId() != "order2" {
		t.Errorf("Expected removed order id 'order2', got %v", removed)
	}
	if q.Length() != 2 {
		t.Errorf("Expected length 2 after remove, got %d", q.Length())
	}

	// Confirm pushing and popping order.
	popped := q.PopNext()
	if popped.GetId() != "order3" {
		t.Errorf("Expected first order id 'order3' after PushFront, got %s", popped.GetId())
	}
}

// --------------------- PriceNodeMap Tests ---------------------
func TestPriceNodeMapPushPop(t *testing.T) {
	pm := NewPriceNodeMap(NewPriceNodeMapParams{NewOrderBookDataStructureParams{}})

	// Create two orders with different prices.
	lowOrder := newFakeOrder("low", 50.0)
	highOrder := newFakeOrder("high", 150.0)

	// Push orders. (Order of push doesn't affect currentBestPrice.)
	pm.Push(lowOrder)
	pm.Push(highOrder)

	// Expect currentBestPrice to be 150.
	if popped := pm.PopNext(); popped.GetId() != "high" {
		t.Errorf("Expected high order to be popped first")
	}

	// After pop, only lowOrder should remain.
	if pm.Length() != 1 {
		t.Errorf("Expected map length 1, got %d", pm.Length())
	}
}

func TestPriceNodeMapPushFrontAndRemove(t *testing.T) {
	pm := NewPriceNodeMap(NewPriceNodeMapParams{NewOrderBookDataStructureParams{}})

	// Create orders at the same price.
	order1 := newFakeOrder("order1", 100.0)
	order2 := newFakeOrder("order2", 100.0)

	// Use Push for order1 and PushFront for order2.
	pm.Push(order1)
	pm.PushFront(order2)

	// Since both are at price 100.0, currentBestPrice is 100.0.
	// PopNext should remove the front of the queue; since order2 was PushFront, expect order2.
	popped := pm.PopNext()
	if popped.GetId() != "order2" {
		t.Errorf("Expected order2 to be popped first, got %s", popped.GetId())
	}

	// Now remove order1 explicitly.
	removed := pm.Remove(RemoveParams{StockId: "order1", priceKey: 100.0})
	if removed == nil || removed.GetId() != "order1" {
		t.Errorf("Failed to remove order1, got %v", removed)
	}

	// With no orders at price 100, Length should be 0.
	if pm.Length() != 0 {
		t.Errorf("Expected map length 0 after removals, got %d", pm.Length())
	}
}

// --------------------- Additional Queue Test Using Direct list ---------------------
// (This test simply confirms that Queue uses container/list as expected.)
func TestQueueInternalStructure(t *testing.T) {
	q := NewQueue(NewQueueParams{NewOrderBookDataStructureParams{}})
	// Use reflection on q.data if needed.
	// For brevity, we check that q.data is initialized.
	if q.data.Len() != 0 {
		t.Errorf("Expected initial list length 0, got %d", q.data.Len())
	}
	// ...existing code...
}
