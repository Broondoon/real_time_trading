package orderExecutorService

// import (
// 	"Shared/network"
// 	"databaseAccessUserManagement"
// 	"encoding/json"
// 	"net/http"
// 	"net/url"
// )

// var _databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface
// var _matchingEngineClientManager network.ClientInterface

// func InitalizeExecutorHandlers(
// 	networkManager network.NetworkInterface,
// 	queueManager network.NetworkInterface,
// 	databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface) {

// 	_databaseAccessUser = databaseAccessUser
// 	_matchingEngineClientManager = queueManager.MatchingEngine()

// 	queueManager.AddHandleFuncUnprotected(network.HandlerParams{Pattern: "executor", Handler: executorHandler})

// 	http.HandleFunc("/health", healthHandler)
// 	queueManager.Listen()
// }

// func healthHandler(w http.ResponseWriter, r *http.Request) {
// 	w.WriteHeader(http.StatusOK)
// 	//log.Println(w, "OK")
// }

// func executorHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
// 	var orderData network.MatchingEngineToExecutionJSON
// 	err := json.Unmarshal(data, &orderData)
// 	if err != nil {
// 		responseWriter.WriteHeader(http.StatusBadRequest)
// 		return
// 	}

// 	// Process the orderData (transferEntity) from the Matching Engine
// 	stockID, buyerStockOrderID, sellerStockOrderID, buySuccess, sellSuccess, err := ProcessTrade(orderData, _databaseAccessUser, _matchingEngineClientManager)
// 	if err != nil {
// 		responseWriter.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}
// 	// Independent failure flags //
// 	// If the match was successful, both IsBuyFailure and IsSellFailure will be false
// 	// If the match was unsuccessful, only one of IsBuyFailure and IsSellFailure will be true
// 	// If there was an error, both IsBuyFailure and IsSellFailure will be true
// 	responseEntity := network.ExecutorToMatchingEngineJSON{
// 		IsBuyFailure:       !buySuccess,
// 		IsSellFailure:      !sellSuccess,
// 		StockID:            stockID,
// 		BuyerStockOrderId:  buyerStockOrderID,
// 		SellerStockOrderId: sellerStockOrderID,
// 	}

// 	jsonResponseToMatchingEngine, err := json.Marshal(responseEntity)
// 	if err != nil {
// 		responseWriter.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}

// 	//responseWriter.Header().Set("Content-Type", "application/json")
// 	// Send the info of whether the match was successful or not
// 	responseWriter.Write(jsonResponseToMatchingEngine)

// }
