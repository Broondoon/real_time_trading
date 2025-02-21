package orderExecutorService


import (
    "Shared/network"
    "Shared/entities/wallet"
    "Shared/entities/transaction"
    "Shared/entities/entity"
    "Shared/entities/user-stock"
    "databaseAccessTransaction"
    "databaseAccessUserManagement"
    "errors"
    "fmt"
    "time"
)

func ProcessTrade(orderData network.MatchingEngineToExecutionJSON, databaseAccessTransact databaseAccessTransaction.DatabaseAccessInterface, databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface) (bool, error) {


    // Transfer Entity
    buyerID := orderData.BuyerID
    sellerID := orderData.SellerID
    stockID := orderData.StockID
    buyOrderID := orderData.BuyOrderID
    sellOrderID := orderData.SellOrderID
    isBuyPartial := orderData.IsBuyPartial
    isSellPartial := orderData.IsSellPartial
    stockPrice := orderData.StockPrice
    quantity := orderData.Quantity

    totalCost := calculateTotalTransactionCost(quantity, stockPrice)


    transactionList, err := databaseAccessTransact.StockTransaction().GetByIDs([]string{buyOrderID, sellOrderID})

    if err != nil {
        return false, fmt.Errorf("failed to get transactions: %v", err)
    }



    walletList, err := databaseAccessUser.Wallet().GetByIDs([]string{buyerID, sellerID})
    if err != nil {
        return false, fmt.Errorf("failed to get wallets: %v", err)
    }
    


    // Check if buyer has enough funds
    buyerHasFunds, err := validateBuyerWalletBalance((*walletList)[0], totalCost)
    if err != nil {
        return false, err
    }
    if !buyerHasFunds {
        return false, nil
    }

    // Get the stock transaction that initiated this trade
    stockTx := (*transactionList)[0]
        
    // Update wallet balances and create wallet transactions
    err = updateWalletBalances(
        (*walletList)[0], // buyer wallet
        (*walletList)[1], // seller wallet
        totalCost,
        stockTx,
        databaseAccessUser,
        databaseAccessTransact,
    )
    if err != nil {
        return false, fmt.Errorf("failed to update wallet balances: %v", err)
    }

    // Update user stock portfolios
    err = updateUserStocks(
        buyerID,
        sellerID,
        stockID,
        quantity,
        stockTx,
        databaseAccessUser,
        databaseAccessTransact,
        isBuyPartial,
        isSellPartial,
    )

    if err != nil {
        return false, fmt.Errorf("failed to update user stocks: %v", err)
    }

    return true, nil
}





func updateWalletBalances(
    buyerWallet wallet.WalletInterface, 
    sellerWallet wallet.WalletInterface, 
    totalCost float64,
    stockTransaction transaction.StockTransactionInterface,
    databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface,
    databaseAccessTransact databaseAccessTransaction.DatabaseAccessInterface) error {


    buyerWT := transaction.NewWalletTransaction(transaction.NewWalletTransactionParams{
        NewEntityParams: entity.NewEntityParams{
            DateCreated:  time.Now(),
            DateModified: time.Now(),
        },
        WalletID:         buyerWallet.GetId(),
        StockTransaction: stockTransaction,
        IsDebit:         true,
        Amount:          totalCost,
    })
    buyerWT.SetTimestamp(time.Now())


    sellerWT := transaction.NewWalletTransaction(transaction.NewWalletTransactionParams{
        NewEntityParams: entity.NewEntityParams{
            DateCreated:  time.Now(),
            DateModified: time.Now(),
        },
        WalletID:         sellerWallet.GetId(),
        StockTransaction: stockTransaction,
        IsDebit:         false,
        Amount:          totalCost,
    })
    sellerWT.SetTimestamp(time.Now())


    buyerWallet.SetBalance(buyerWallet.GetBalance() - totalCost)
    err := databaseAccessUser.Wallet().Update(buyerWallet)
    if err != nil {
        return fmt.Errorf("failed to update buyer wallet: %v", err)
    }

    sellerWallet.SetBalance(sellerWallet.GetBalance() + totalCost)
    err = databaseAccessUser.Wallet().Update(sellerWallet)
    if err != nil {
        return fmt.Errorf("failed to update seller wallet: %v", err)
    }

    // Use the generic Create function from EntityDataAccessInterface
    _, err = databaseAccessTransact.WalletTransaction().Create(buyerWT)
    if err != nil {
        return fmt.Errorf("failed to create buyer wallet transaction: %v", err)
    }

    _, err = databaseAccessTransact.WalletTransaction().Create(sellerWT)
    if err != nil {
        return fmt.Errorf("failed to create seller wallet transaction: %v", err)
    }

    return nil
}






func updateUserStocks(
    buyerID string,
    sellerID string,
    stockID string,
    quantity int,
    stockTx transaction.StockTransactionInterface,
    databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface,
    databaseAccessTransact databaseAccessTransaction.DatabaseAccessInterface,
    isBuyPartial bool,
    isSellPartial bool) error {

    // Get buyer's current stock holdings
    buyerStocks, err := databaseAccessUser.UserStock().GetUserStocks(buyerID)
    if err != nil {
        return fmt.Errorf("failed to get buyer stocks: %v", err)
    }

    // Get seller's current stock holdings
    sellerStocks, err := databaseAccessUser.UserStock().GetUserStocks(sellerID)
    if err != nil {
        return fmt.Errorf("failed to get seller stocks: %v", err)
    }

    var sellerStock userStock.UserStockInterface
    for _, stock := range *sellerStocks {
        if stock.GetStockID() == stockID {
            sellerStock = stock
            break
        }
    }

    if sellerStock == nil {
        return fmt.Errorf("seller does not own stock %s", stockID)
    }


    if sellerStock.GetQuantity() < quantity {
        return fmt.Errorf("seller does not have enough shares of stock %s", stockID)
    }


    var buyerStock userStock.UserStockInterface
    for _, stock := range *buyerStocks {
        if stock.GetStockID() == stockID {
            buyerStock = stock
            break
        }
    }

    if buyerStock == nil {
        buyerStock = userStock.New(userStock.NewUserStockParams{
            NewEntityParams: entity.NewEntityParams{
                DateCreated:  time.Now(),
                DateModified: time.Now(),
            },
            UserID:   buyerID,
            StockID:  stockID,
            Quantity: 0,
        })
        
        // Create in database first
        createdStock, err := databaseAccessUser.UserStock().Create(buyerStock)
        if err != nil {
            return fmt.Errorf("failed to create buyer stock holding: %v", err)
        }
        buyerStock = createdStock
    }

    // Check if this is a complete or partial fill using flags from matching engine
    if !isBuyPartial && !isSellPartial {
        // Both orders are complete
        stockTx.SetOrderStatus("COMPLETE")
        err = databaseAccessTransact.StockTransaction().Update(stockTx)
        if err != nil {
            return fmt.Errorf("failed to update stock transaction status: %v", err)
        }
    } else {
        // At least one order is partial
        stockTx.SetOrderStatus("PARTIAL")
        err = databaseAccessTransact.StockTransaction().Update(stockTx)
        if err != nil {
            return fmt.Errorf("failed to update original stock transaction status: %v", err)
        }

        // Create new transaction for the filled portion
        filledTx := transaction.NewStockTransaction(transaction.NewStockTransactionParams{
            NewEntityParams: entity.NewEntityParams{
                DateCreated:  time.Now(),
                DateModified: time.Now(),
            },
            StockID:                  stockID,
            ParentStockTransactionID: stockTx.GetId(),
            OrderStatus:              "COMPLETE",
            IsBuy:                    stockTx.GetIsBuy(),
            OrderType:                stockTx.GetOrderType(),
            StockPrice:               stockTx.GetStockPrice(),
            Quantity:                 quantity,
        })
        filledTx.SetTimestamp(time.Now())

        _, err = databaseAccessTransact.StockTransaction().Create(filledTx)
        if err != nil {
            return fmt.Errorf("failed to create filled stock transaction: %v", err)
        }
    }

    // Update quantities
    buyerStock.SetQuantity(buyerStock.GetQuantity() + quantity)
    sellerStock.SetQuantity(sellerStock.GetQuantity() - quantity)

    // Update in database
    err = databaseAccessUser.UserStock().Update(buyerStock)
    if err != nil {
        return fmt.Errorf("failed to update buyer stock: %v", err)
    }

    err = databaseAccessUser.UserStock().Update(sellerStock)
    if err != nil {
        return fmt.Errorf("failed to update seller stock: %v", err)
    }

    return nil
}




func calculateTotalTransactionCost(quantity int, stockPrice float64) float64 {
    return float64(quantity) * stockPrice
}




func validateBuyerWalletBalance(buyerWallet wallet.WalletInterface, totalCost float64) (bool, error) {
    if buyerWallet == nil {
        return false, errors.New("buyer wallet not found")
    }

    buyerBalance := buyerWallet.GetBalance()
    return buyerBalance >= totalCost, nil
}

