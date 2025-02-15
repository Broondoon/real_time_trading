package matchingEngine

// //Created using Copilot and ChatGPT 03-mini (preview)
// //TODO: FIX TESTS and ensure full code coverage.
// import (
// 	"encoding/json"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"Shared/entities/stock"
// 	"Shared/network"
// )

// func TestAddNewStockHandler_Valid(t *testing.T) {
// 	// Reset global map.
// 	_matchingEngineMap = make(map[string]MatchingEngineInterface)

// 	// Create a valid JSON.
// 	// Assume entity.NewEntityParams includes an "Id" field.
// 	input := map[string]interface{}{
// 		"Name": "TestStock",
// 		"NewEntityParams": map[string]interface{}{
// 			"Id": "stock123",
// 		},
// 	}
// 	jsonData, _ := json.Marshal(input)

// 	rec := httptest.NewRecorder()
// 	AddNewStockHandler(rec, jsonData)

// 	// Check that the stock is added in the global map.
// 	if _, exists := _matchingEngineMap["stock123"]; !exists {
// 		t.Errorf("Expected stock with id 'stock123' in _matchingEngineMap")
// 	}
// 	// No bad status expected.
// 	if rec.Code == http.StatusBadRequest {
// 		t.Errorf("Unexpected status %d", rec.Code)
// 	}
// }

// func TestAddNewStockHandler_InvalidJSON(t *testing.T) {
// 	// Reset global map.
// 	_matchingEngineMap = make(map[string]MatchingEngineInterface)

// 	rec := httptest.NewRecorder()
// 	invalidJSON := []byte("invalid json")
// 	AddNewStockHandler(rec, invalidJSON)

// 	// Expect BadRequest when parse fails.
// 	if rec.Code != http.StatusBadRequest {
// 		t.Errorf("Expected status %d; got %d", http.StatusBadRequest, rec.Code)
// 	}
// }

// func TestInitializeHandlers(t *testing.T) {
// 	// Reset global map.
// 	_matchingEngineMap = make(map[string]MatchingEngineInterface)

// 	// Create a slice with one fake stock.
// 	fakeEntity := stock.FakeStock{
// 		Name: "FakeStock",
// 	}
// 	// We need a minimal implementation of GetId.
// 	// For the purpose of testing, assume FakeStock implements GetId by returning a fixed id.
// 	fakeStock := struct {
// 		stock.StockInterface
// 	}{
// 		StockInterface: &fakeEntity,
// 	}
// 	stockList := []stock.StockInterface{fakeStock}

// 	// Use fake network manager.
// 	fakeNM := &network.FakeNetworkManager{}

// 	// Call the initializer.
// 	InitalizeHandlers(stockList, fakeNM)

// 	// Check that Listen was called with the proper port.
// 	if !fakeNM.ListenCalled {
// 		t.Error("Expected Listen to be called")
// 	}
// 	if fakeNM.ListenerParams.Port != "8080" {
// 		t.Errorf("Expected port '8080'; got '%s'", fakeNM.ListenerParams.Port)
// 	}
// }
