package orderExecutorService

import (
    "Shared/network"
    "databaseAccessTransaction"
    "databaseAccessUserManagement"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
)

var databaseAccessTransact databaseAccessTransaction.DatabaseAccessInterface
var databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface



func InitalizeExecutorHandlers(
    networkManager network.NetworkInterface, 
    databaseAccessTransact databaseAccessTransaction.DatabaseAccessInterface, 
    databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface) {

    networkManager.AddHandleFuncUnprotected(network.HandlerParams{Pattern: "executor", Handler: executorHandler})

    http.HandleFunc("/health", healthHandler)
}



func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    fmt.Fprintln(w, "OK")
}




func executorHandler(responseWriter http.ResponseWriter, data []byte, queryParams url.Values, requestType string) {
    var orderData network.MatchingEngineToExecutionJSON
    err := json.Unmarshal(data, &orderData)
    if err != nil {
        responseWriter.WriteHeader(http.StatusBadRequest)
        return
    }

    // Process the orderData (transferEntity) from the Matching Engine
    buySuccess, sellSuccess, err := ProcessTrade(orderData, databaseAccessTransact, databaseAccessUser)
    if err != nil {
        responseWriter.WriteHeader(http.StatusInternalServerError)
        return
    }


    // Independent failure flags //
    // If the match was successful, both IsBuyFailure and IsSellFailure will be false
    // If the match was unsuccessful, only one of IsBuyFailure and IsSellFailure will be true 
    // If there was an error, both IsBuyFailure and IsSellFailure will be true
    responseEntity := network.ExecutorToMatchingEngineJSON{
        IsBuyFailure:  !buySuccess,
        IsSellFailure: !sellSuccess,
    }


    jsonResponseToMatchingEngine, err := json.Marshal(responseEntity)
    if err != nil {
        responseWriter.WriteHeader(http.StatusInternalServerError)
        return
    }

    //responseWriter.Header().Set("Content-Type", "application/json")



    // Send a 200 OK response if the request was successful (unrelated to match success)
    responseWriter.WriteHeader(http.StatusOK)


    // Send the info of whether the match was successful or not
    responseWriter.Write(jsonResponseToMatchingEngine)

}