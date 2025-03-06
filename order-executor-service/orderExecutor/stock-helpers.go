package orderExecutorService

import (
    //"Shared/entities/entity"
    "Shared/entities/transaction"
    userStock "Shared/entities/user-stock"
    "databaseAccessTransaction"
    "databaseAccessUserManagement"
    "fmt"
    "time"
)




// Calculates the total cost of a transaction given the quantity and stock price.
func calculateTotalTransactionCost(quantity int, stockPrice float64) float64 {
    return float64(quantity) * stockPrice
}




// Finds and validates user stock portfolios
func findUserStockPortfolios(
    buyerID string,
    sellerID string,
    databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface,
) (*[]userStock.UserStockInterface, *[]userStock.UserStockInterface, error) {

    // Get buyer's current stock holdings
    buyerStockPortfolio, err := databaseAccessUser.UserStock().GetUserStocks(buyerID)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to get buyer stocks: %v", err)
    }
    println("Retrieved buyer portfolio with %d stocks", len(*buyerStockPortfolio))

    // Get seller's current stock holdings
    sellerStockPortfolio, err := databaseAccessUser.UserStock().GetUserStocks(sellerID)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to get seller stocks: %v", err)
    }
    println("Retrieved seller portfolio with %d stocks", len(*sellerStockPortfolio))

    return buyerStockPortfolio, sellerStockPortfolio, nil
}




// Finds and validates seller's stock holding
func handleSellerStock(
    sellerStockPortfolio *[]userStock.UserStockInterface,
    stockID string,
    quantity int,
) (userStock.UserStockInterface, error) {

    var sellerStock userStock.UserStockInterface
    for _, stock := range *sellerStockPortfolio {
        if stock.GetStockID() == stockID {
            sellerStock = stock
            break
        }
    }

    if sellerStock == nil {
        return nil, fmt.Errorf("seller does not own stock %s", stockID)
    }

    if sellerStock.GetQuantity() < quantity {
        return nil, fmt.Errorf("seller does not have enough shares of stock %s", stockID)
    }

    println("Initial quantities - Seller: %d", sellerStock.GetQuantity())
    return sellerStock, nil
}





// Creates or retrieves buyer's stock holding
func handleBuyerStock(
    buyerStockPortfolio *[]userStock.UserStockInterface,
    buyerID string,
    stockID string,
    //quantity int,
    sellerStock userStock.UserStockInterface,
    databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface,
) (userStock.UserStockInterface, error) {

    var buyerStock userStock.UserStockInterface
    for _, stock := range *buyerStockPortfolio {
        if stock.GetStockID() == stockID {
            buyerStock = stock
            break
        }
    }
    println("Buyer initial: %d", buyerStock.GetQuantity())

    if buyerStock == nil {
        buyerStock = userStock.New(userStock.NewUserStockParams{
            UserID:    buyerID,
            StockID:   stockID,
            StockName: sellerStock.GetStockName(),
            Quantity:  0,
        })
        createdStock, err := databaseAccessUser.UserStock().Create(buyerStock)
        if err != nil {
            return nil, fmt.Errorf("failed to create buyer stock holding: %v", err)
        }
        buyerStock = createdStock
    }
    return buyerStock, nil
}




// Updates stock quantities in database
func updateUserStockQuantities(
    buyerStock userStock.UserStockInterface,
    sellerStock userStock.UserStockInterface,
    quantity int,
    databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface,
) error {

    sellerStock.SetQuantity(sellerStock.GetQuantity() - quantity)
    buyerStock.SetQuantity(buyerStock.GetQuantity() + quantity)

    if err := databaseAccessUser.UserStock().Update(sellerStock); err != nil {
        return fmt.Errorf("failed to update seller stock: %v", err)
    }

    if err := databaseAccessUser.UserStock().Update(buyerStock); err != nil {
        return fmt.Errorf("failed to update buyer stock: %v", err)
    }

    println("Final quantities - Buyer: %d, Seller: %d", buyerStock.GetQuantity(), sellerStock.GetQuantity())

    return nil
}





// Updates transaction status and creates filled transaction if needed
// In order-executor-service/orderExecutor/stock-helpers.go
func updateTransactionStatus(
    stockTx transaction.StockTransactionInterface,
    isBuyPartial bool,
    isSellPartial bool,
    stockPrice float64,
    databaseAccessTransact databaseAccessTransaction.DatabaseAccessInterface,
) error {


    // Handle partial matching for both buy and sell orders
    if stockTx.GetIsBuy() {
        if isBuyPartial {
            stockTx.SetOrderStatus("PARTIALLY_COMPLETE")
        } else {
            stockTx.SetOrderStatus("COMPLETED")
        }
    } else {
        // For sell orders
        if isSellPartial {
            stockTx.SetOrderStatus("PARTIALLY_COMPLETE")
        } else {
            stockTx.SetOrderStatus("COMPLETED")
        }
    }
    
    // Set the stock price in the transaction
    stockTx.SetStockPrice(stockPrice)
    
    // Update in database
    if err := databaseAccessTransact.StockTransaction().Update(stockTx); err != nil {
        return fmt.Errorf("failed to update transaction status: %v", err)
    }

    // Create filled transaction for partial orders
    if isBuyPartial || isSellPartial {
        filledTx := transaction.NewStockTransaction(transaction.NewStockTransactionParams{
            ParentStockTransaction: stockTx,
            OrderStatus:            "COMPLETED", // Child transaction is always COMPLETED
            TimeStamp:              time.Now(),
        })
        
        // Set the stock price in the filled transaction
        filledTx.SetStockPrice(stockPrice)

        if _, err := databaseAccessTransact.StockTransaction().Create(filledTx); err != nil {
            return fmt.Errorf("failed to create filled stock transaction: %v", err)
        }
    }
    
    return nil
}