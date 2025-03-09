package orderExecutorService

import (
	"Shared/network"
	"databaseAccessTransaction"
	"databaseAccessUserManagement"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

var _databaseAccessTransact databaseAccessTransaction.DatabaseAccessInterface
var _databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface

func InitalizeExecutorHandlers(
	networkManager network.NetworkInterface,
	databaseAccessTransact databaseAccessTransaction.DatabaseAccessInterface,
	databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface) {

	_databaseAccessTransact = databaseAccessTransact
	_databaseAccessUser = databaseAccessUser

	networkManager.AddHandleFuncUnprotected(network.HandlerParams{Pattern: "executor", Handler: executorHandler})

	http.HandleFunc("/health", healthHandler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	log.Println(w, "OK")
}

func executorHandler(responseWriter network.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
	var orderData network.MatchingEngineToExecutionJSON
	err := json.Unmarshal(data, &orderData)
	if err != nil {
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}

	// Process the orderData (transferEntity) from the Matching Engine
	buySuccess, sellSuccess, err := ProcessTrade(orderData, _databaseAccessTransact, _databaseAccessUser)
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Println(fmt.Sprintf("Done ProcessTrade - buySuccess: %t, sellSuccess: %t", buySuccess, sellSuccess))
	// Independent failure flags //
	// If the match was successful, both IsBuyFailure and IsSellFailure will be false
	// If the match was unsuccessful, only one of IsBuyFailure and IsSellFailure will be true
	// If there was an error, both IsBuyFailure and IsSellFailure will be true
	responseEntity := network.ExecutorToMatchingEngineJSON{
		IsBuyFailure:  !buySuccess,
		IsSellFailure: !sellSuccess,
	}

	log.Println(fmt.Sprintf("IsBuyFailure: %t, IsSellFailure: %t", responseEntity.IsBuyFailure, responseEntity.IsSellFailure))
	jsonResponseToMatchingEngine, err := json.Marshal(responseEntity)
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	//responseWriter.Header().Set("Content-Type", "application/json")
	// Send the info of whether the match was successful or not
	responseWriter.Write(jsonResponseToMatchingEngine)

}
