package orderExecutorService

import (
	//"Shared/entities/entity"
	"Shared/entities/transaction"
	userStock "Shared/entities/user-stock"
	"databaseAccessTransaction"
	"databaseAccessUserManagement"
	"fmt"

	"github.com/google/uuid"
)

// Calculates the total cost of a transaction given the quantity and stock price.
func calculateTotalTransactionCost(quantity int, stockPrice float64) float64 {
	return float64(quantity) * stockPrice
}

// Finds and validates user stock portfolios
func findUserStockPortfolios(
	buyerID *uuid.UUID,
	sellerID *uuid.UUID,
	databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface,
) (*[]userStock.UserStockInterface, error) {

	// Get buyer's current stock holdings
	buyerStockPortfolio, err := databaseAccessUser.UserStock().GetUserStocks(buyerID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get buyer stocks: %v", err)
	}
	//log.Printf("Retrieved buyer portfolio with %d stocks", len(*buyerStockPortfolio))

	// // Get seller's current stock holdings
	// sellerStockPortfolio, err := databaseAccessUser.UserStock().GetUserStocks(sellerID.String())
	// if err != nil {
	// 	return nil, nil, fmt.Errorf("failed to get seller stocks: %v", err)
	// }
	// //log.Printf("%s", fmt.Sprintf("Retrieved seller portfolio with %d stocks", len(*sellerStockPortfolio)))

	return buyerStockPortfolio, nil
}

// Finds and validates seller's stock holding
func handleSellerStock(
	sellerStockPortfolio *[]userStock.UserStockInterface,
	stockID *uuid.UUID,
	quantity int,
) (userStock.UserStockInterface, error) {

	var sellerStock userStock.UserStockInterface
	for _, stock := range *sellerStockPortfolio {
		if stock.GetStockIDString() == stockID.String() {
			sellerStock = stock
			break
		}
	}

	return sellerStock, nil
	//println(fmt.Sprintf("Final -> Seller  has %d shares of StockID: %s", sellerStock.GetQuantity(), sellerStock.GetStockID()))
}

// Creates or retrieves buyer's stock holding
func handleBuyerStock(
	buyerStockPortfolio *[]userStock.UserStockInterface,
	buyerID *uuid.UUID,
	stockID *uuid.UUID,
	//quantity int,
	stockName string,
	databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface,
) (userStock.UserStockInterface, error) {

	var buyerStock userStock.UserStockInterface
	for _, stock := range *buyerStockPortfolio {
		if stock.GetStockIDString() == stockID.String() {
			buyerStock = stock
			break
		}
	}
	// If the buyer doesn't have any of the stock, a new stock holding is created
	// The quantity is originally set to zero and is updated after (otherwise there's an error where the quantity is double what it should be)

	if buyerStock == nil {
		buyerStock = userStock.New(userStock.NewUserStockParams{
			UserID:    buyerID,
			StockID:   stockID,
			StockName: stockName,
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
	quantity int,
	databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface,
) error {

	buyerStock.UpdateQuantity(quantity)

	if err := databaseAccessUser.UserStock().Update(buyerStock); err != nil {
		return fmt.Errorf("failed to update buyer stock: %v", err)
	}
	//log.Println(fmt.Sprintf("Final -> Buyer  has %d shares of StockID: %s", buyerStock.GetQuantity(), buyerStock.GetStockID()))

	//log.Printf("Final Buyer Quantity of StockID = %s is %d, Final Seller Quantity of StockID = %s is %d", buyerStock.GetStockID(), buyerStock.GetQuantity(), sellerStock.GetStockID(), sellerStock.GetQuantity())

	return nil
}

// Updates transaction status and creates filled transaction if needed
func updateTransactionStatus(
	buyerStockTx transaction.StockTransactionInterface,
	buyerWalletTx transaction.WalletTransactionInterface,
	sellerStockTx transaction.StockTransactionInterface,
	sellerWalletTx transaction.WalletTransactionInterface,
	isBuyPartial bool,
	isSellPartial bool,
	stockPrice float64,
	databaseAccessTransact databaseAccessTransaction.DatabaseAccessInterface,
	quantity int,
) error {

	buyerStockTx.UpdateStockPrice(stockPrice)
	// Handle partial matching for both buy and sell orders
	updateTransaction(buyerStockTx, isBuyPartial, stockPrice, databaseAccessTransact, quantity, buyerWalletTx)
	updateTransaction(sellerStockTx, isSellPartial, stockPrice, databaseAccessTransact, quantity, sellerWalletTx)

	// Update in database

	return nil
}

func updateTransaction(
	stockTx transaction.StockTransactionInterface,
	isPartial bool,
	stockPrice float64,
	databaseAccessTransact databaseAccessTransaction.DatabaseAccessInterface,
	quantity int,
	walletTx transaction.WalletTransactionInterface,
) error {
	if stockTx.GetOrderStatus() == "PARTIALLY_COMPLETE" {
		partials, err := _databaseAccessTransact.StockTransaction().GetByForeignID("ParentStockTransactionID", stockTx.GetIdString())
		if err != nil {
			return fmt.Errorf("failed to get partial transactions: %v", err)
		}
		quantityTransfered := 0
		for _, partial := range *partials {
			quantityTransfered += partial.GetQuantity()
		}
		if quantityTransfered+quantity == stockTx.GetQuantity() {
			stockTx.SetOrderStatus("COMPLETED")
		}
	} else if isPartial {
		stockTx.SetOrderStatus("PARTIALLY_COMPLETE")
	} else {
		stockTx.SetOrderStatus("COMPLETED")
		stockTx.SetWalletTransactionID(walletTx.GetId())
	}

	if err := databaseAccessTransact.StockTransaction().Update(stockTx); err != nil {
		return fmt.Errorf("failed to update transaction status: %v", err)
	}
	//log.Printf("%s", fmt.Sprintf("AFTER Update Status: %s", stockTx.GetOrderStatus()))

	// Create filled transaction for partial orders
	if isPartial {
		var filledTx transaction.StockTransactionInterface = transaction.NewStockTransaction(transaction.NewStockTransactionParams{
			ParentStockTransaction: stockTx,
		})

		// Set the stock price in the filled transaction
		filledTx.UpdateStockPrice(stockPrice)
		filledTx.SetOrderStatus("COMPLETED")
		filledTx.SetWalletTransactionID(walletTx.GetId())
		filledTx, err := databaseAccessTransact.StockTransaction().Create(filledTx)
		if err != nil {
			return fmt.Errorf("failed to create filled stock transaction: %v", err)
		}

		walletTx.SetStockTransactionID(filledTx.GetId())
		if err := databaseAccessTransact.WalletTransaction().Update(walletTx); err != nil {
			return fmt.Errorf("failed to update wallet transaction: %v", err)
		}

		//	log.Printf("%s", fmt.Sprintf("Created Filled Transaction with ID: %s", filledTx.GetId()))
	}

	return nil
}
