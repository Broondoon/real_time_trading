package main

import (
    OrderExecutorService "OrderExecutorService/orderExecutor"
    "Shared/network"
    "databaseAccessTransaction"
    "databaseAccessUserManagement"
)

func main() {
    networkManager := network.NewNetwork()
    databaseAccessTransaction := databaseAccessTransaction.NewDatabaseAccess(&databaseAccessTransaction.NewDatabaseAccessParams{
        Network: networkManager,
    })
    databaseAccessUserManagement := databaseAccessUserManagement.NewDatabaseAccess(&databaseAccessUserManagement.NewDatabaseAccessParams{
        Network: networkManager,
    })

    go OrderExecutorService.InitalizeExecutorHandlers(networkManager, databaseAccessTransaction, databaseAccessUserManagement)
    println("Order Executor Service Started")

    networkManager.Listen(network.ListenerParams{
        Handler: nil,
    })
}
