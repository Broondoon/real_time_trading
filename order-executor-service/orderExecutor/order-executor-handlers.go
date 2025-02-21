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
var networkManager network.NetworkInterface

func InitalizeExecutorHandlers(
    networkManager network.NetworkInterface, databaseAccessTransact databaseAccessTransaction.DatabaseAccessInterface, databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface) {
    databaseAccessTransact = databaseAccessTransact
    databaseAccessUser = databaseAccessUser
    networkManager = networkManager

    // Build Handler
    networkManager.AddHandleFuncUnprotected(network.HandlerParams{Pattern: "executor", Handler: executorHandler})

    http.HandleFunc("/health", healthHandler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    fmt.Fprintln(w, "OK")
}
