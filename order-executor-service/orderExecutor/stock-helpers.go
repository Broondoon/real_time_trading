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
	println(fmt.Sprintf("Retrieved buyer portfolio with %d stocks", len(*buyerStockPortfolio)))

	// Get seller's current stock holdings
	sellerStockPortfolio, err := databaseAccessUser.UserStock().GetUserStocks(sellerID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get seller stocks: %v", err)
	}
	println(fmt.Sprintf("Retrieved seller portfolio with %d stocks", len(*sellerStockPortfolio)))

	return buyerStockPortfolio, sellerStockPortfolio, nil
}

// Finds and validates seller's stock holding
func handleSellerStock(
	sellerStockPortfolio *[]userStock.UserStockInterface,
	stockID string,
	quantity int,
) (userStock.UserStockInterface, error) {

	var sellerStock userStock.UserStockInterface

	var count int
	for _, stock := range *sellerStockPortfolio {
		//(stock.ToParams().StockID)
		//println(stock.ToParams().StockName)
		//println(stock.ToParams().Quantity)
		//count++
		println(fmt.Sprintf("Portfolio Stock: %s, Portfolio Stock Quantity: %d", stock.GetStockID(), stock.GetQuantity()))
		if stock.GetStockID() == stockID {
			sellerStock = stock
			break
		}
	}
	println(count)

	//println(fmt.Sprintf("Final -> Seller  has %d shares of StockID: %s", sellerStock.GetQuantity(), sellerStock.GetStockID()))

	sellerQuantity := sellerStock.GetQuantity()

	println(fmt.Sprintf("Initially Seller has %d shares of StockID: %s", sellerQuantity, sellerStock.GetStockID()))
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
	//println(fmt.Sprintf("Initially Buyer has %d shares of StockID: %s", buyerStock.GetQuantity(), buyerStock.GetStockID()))

	// If the buyer doesn't have any of the stock, a new stock holding is created
	// The quantity is originally set to zero and is updated after (otherwise there's an error where the quantity is double what it should be)
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

// Updates the user's stock quantities in the database
func updateUserStockQuantities(
	buyerStock userStock.UserStockInterface,
	sellerStock userStock.UserStockInterface,
	quantity int,
	databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface,
) error {

	//sellerStock.SetQuantity(sellerStock.GetQuantity() - quantity)
	buyerStock.SetQuantity(buyerStock.GetQuantity() + quantity)

	if err := databaseAccessUser.UserStock().Update(sellerStock); err != nil {
		return fmt.Errorf("failed to update seller stock: %v", err)
	}
	//println(fmt.Sprintf("Final -> Seller  has %d shares of StockID: %s", sellerStock.GetQuantity(), sellerStock.GetStockID()))

	if err := databaseAccessUser.UserStock().Update(buyerStock); err != nil {
		return fmt.Errorf("failed to update buyer stock: %v", err)
	}
	//println(fmt.Sprintf("Final -> Buyer  has %d shares of StockID: %s", buyerStock.GetQuantity(), buyerStock.GetStockID()))

	println("Final Buyer Quantity of StockID = %s is %d, Final Seller Quantity of StockID = %s is %d", buyerStock.GetStockID(), buyerStock.GetQuantity(), sellerStock.GetStockID(), sellerStock.GetQuantity())

	return nil
}

// Updates transaction status and creates filled transaction if needed
func updateTransactionStatus(
	stockTx transaction.StockTransactionInterface,
	isBuyPartial bool,
	isSellPartial bool,
	stockPrice float64,
	databaseAccessTransact databaseAccessTransaction.DatabaseAccessInterface,
) error {

	println(fmt.Sprintf("BEFORE Update Status: %s", stockTx.GetOrderStatus()))

	// Set the stock price in the transaction

	// Handle partial matching for both buy and sell orders
	if stockTx.GetIsBuy() {
		stockTx.SetStockPrice(stockPrice + stockTx.GetStockPrice())
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

	// Update in database
	if err := databaseAccessTransact.StockTransaction().Update(stockTx); err != nil {
		return fmt.Errorf("failed to update transaction status: %v", err)
	}
	println(fmt.Sprintf("AFTER Update Status: %s", stockTx.GetOrderStatus()))

	// Create filled transaction for partial orders
	if isBuyPartial || isSellPartial {
		filledTx := transaction.NewStockTransaction(transaction.NewStockTransactionParams{
			ParentStockTransaction: stockTx,
			//OrderStatus:            "COMPLETED", // Child transaction is always COMPLETED
			//TimeStamp:              time.Now(),
		})

		// Set the stock price in the filled transaction
		//filledTx.SetStockPrice(stockPrice)

		filledTx.SetOrderStatus("COMPLETED")
		filledTx.SetStockPrice(stockPrice)
		filledTx.SetTimestamp(time.Now())
		if _, err := databaseAccessTransact.StockTransaction().Create(filledTx); err != nil {
			return fmt.Errorf("failed to create filled stock transaction: %v", err)
		}

		println(fmt.Sprintf("Created Filled Transaction with ID: %s", filledTx.GetId()))
	}

	return nil
}
