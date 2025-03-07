package orderExecutorService

import (
	"Shared/entities/entity"
	"Shared/entities/transaction"
	userStock "Shared/entities/user-stock"
	"Shared/entities/wallet"
	"Shared/network"
	"databaseAccessTransaction"
	"databaseAccessUserManagement"
	"errors"
	"fmt"
	"time"
)

// ProcessTrade
func ProcessTrade(orderData network.MatchingEngineToExecutionJSON, databaseAccessTransact databaseAccessTransaction.DatabaseAccessInterface, databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface) (bool, bool, error) {
	//return true, true, nil
	// Transfer Entity received from the Matching Engine //
	buyerID := orderData.BuyerID
	sellerID := orderData.SellerID
	stockID := orderData.StockID
	buyOrderID := orderData.BuyOrderID
	sellOrderID := orderData.SellOrderID
	isBuyPartial := orderData.IsBuyPartial
	isSellPartial := orderData.IsSellPartial
	stockPrice := orderData.StockPrice
	quantity := orderData.Quantity

	println("Buyer ID: ", buyerID)
	println("Seller ID: ", sellerID)
	println("Stock ID: ", stockID)
	println("Buy Order ID: ", buyOrderID)
	println("Sell Order ID: ", sellOrderID)
	println("Is Buy Partial: ", isBuyPartial)
	println("Is Sell Partial: ", isSellPartial)
	println("Stock Price: ", stockPrice)
	println("Quantity: ", quantity)

	totalCost := calculateTotalTransactionCost(quantity, stockPrice)
	println("Total Cost: ", totalCost)

	// 1. Go to the Transaction DB, get any stock transaction with the ID Equal to the buy order ID or the sell order ID
	transactionList, err := databaseAccessTransact.StockTransaction().GetByIDs([]string{buyOrderID, sellOrderID})

	if err != nil {
		println("Error: ", err.Error())
		return false, false, fmt.Errorf("failed to get transactions: %v", err)
	}

	stockTx := (*transactionList)[0] // Get the stock transaction that initiated this trade

	// 2. Go to User-Managment DB, get wallet of userID present on  Buy order transaction
	walletList, err := databaseAccessUser.Wallet().GetByIDs([]string{buyerID, sellerID})
	if err != nil {
		println("Error: ", err.Error())
		return false, false, fmt.Errorf("failed to get wallets: %v", err)
	}

	// 3. Check if buyer has enough funds to afford the quantity*stockprice
	buyerHasFunds, err := validateBuyerWalletBalance((*walletList)[0], totalCost)
	if err != nil {
		println("Error: ", err.Error())
		return false, false, err
	}
	if !buyerHasFunds {
		return false, true, nil
	}

	// 4. Update buyer and seller wallet balances and create wallet transactions for these changes
	err = updateWalletBalances((*walletList)[0], (*walletList)[1], totalCost, stockTx, databaseAccessUser, databaseAccessTransact)
	if len(*walletList) != 2 {
		return false, false, fmt.Errorf("expected 2 wallets, got %d", len(*walletList))
	}
	if err != nil {
		println("Error: ", err.Error())
		return false, false, fmt.Errorf("failed to update wallet balances: %v", err)
	}

	// 5. Update buyer and seller stock portfolios. Deduct the stock quantity from the seller's portfolio.
	//    Add the stock quantity to the buyer's portfolio.
	err = updateUserStocks(buyerID, sellerID, stockID, quantity, stockTx, databaseAccessUser, databaseAccessTransact, isBuyPartial, isSellPartial)
	if err != nil {
		println("Error: ", err.Error())
		if err.Error() == "seller does not have enough shares of stock "+stockID {
			return true, false, nil // Buy succeeds, sell fails
		}
		return false, false, fmt.Errorf("failed to update user stocks: %v", err)
	}

	// 6. Return true to the matching engine to indicate that the trade was successful.
	return true, true, nil
}

// Updates buyer and seller wallet balances and create wallet transactions for these changes
func updateWalletBalances(
	buyerWallet wallet.WalletInterface,
	sellerWallet wallet.WalletInterface,
	totalCost float64,
	stockTransaction transaction.StockTransactionInterface,
	databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface,
	databaseAccessTransact databaseAccessTransaction.DatabaseAccessInterface) error {

	// 1. Update wallet balances first
	buyerWallet.UpdateBalance(-totalCost)
	err := databaseAccessUser.Wallet().Update(buyerWallet)
	if err != nil {
		println("Error: ", err.Error())
		return fmt.Errorf("failed to update buyer wallet balance: %v", err)
	}

	sellerWallet.UpdateBalance(totalCost)
	err = databaseAccessUser.Wallet().Update(sellerWallet)
	if err != nil {
		println("Error: ", err.Error())
		// Rollback buyer's wallet change if seller update fails
		buyerWallet.UpdateBalance(totalCost)

		if rollbackErr := databaseAccessUser.Wallet().Update(buyerWallet); rollbackErr != nil {
			return fmt.Errorf("failed to update seller wallet balance and rollback failed: %v, rollback error: %v", err, rollbackErr)
		}

		return fmt.Errorf("failed to update seller wallet balance: %v", err)
	}

	// Create and save buyer's wallet transaction
	buyerWT := transaction.NewWalletTransaction(transaction.NewWalletTransactionParams{
		NewEntityParams: entity.NewEntityParams{
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		Wallet:           buyerWallet,      // wallet interface
		StockTransaction: stockTransaction, // stock transaction interface
		IsDebit:          true,
		Amount:           totalCost,
		Timestamp:        time.Now(),
	})

	_, err = databaseAccessTransact.WalletTransaction().Create(buyerWT)
	if err != nil {
		println("Error: ", err.Error())
		return fmt.Errorf("failed to create buyer wallet transaction: %v", err)
	}

	// Create and save seller's wallet transaction
	sellerWT := transaction.NewWalletTransaction(transaction.NewWalletTransactionParams{
		NewEntityParams: entity.NewEntityParams{
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		Wallet:           sellerWallet,     // wallet interface
		StockTransaction: stockTransaction, // stock transaction interface
		IsDebit:          false,
		Amount:           totalCost,
		Timestamp:        time.Now(),
	})

	_, err = databaseAccessTransact.WalletTransaction().Create(sellerWT)
	if err != nil {
		println("Error: ", err.Error())
		return fmt.Errorf("failed to create seller wallet transaction: %v", err)
	}

	return nil
}

// Updates buyer and seller stock portfolios following a successful trade
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
	buyerStockPortfolio, err := databaseAccessUser.UserStock().GetUserStocks(buyerID)
	if err != nil {
		if err.Error() == "server returned error: 404 Not Found" {
			buyerStockPortfolio = &[]userStock.UserStockInterface{}

		} else {
			println("Error: ", err.Error())
			return fmt.Errorf("failed to get buyer stocks: %v", err)
		}
	}

	// Get seller's current stock holdings
	sellerStockPortfolio, err := databaseAccessUser.UserStock().GetUserStocks(sellerID)
	if err != nil {
		println("Error: ", err.Error())
		return fmt.Errorf("failed to get seller stocks: %v", err)
	}

	// Find the stock in the seller's portfolio
	var sellerStock userStock.UserStockInterface
	for _, stock := range *sellerStockPortfolio {
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

	// Find the stock in the buyer's portfolio
	var buyerStock userStock.UserStockInterface
	for _, stock := range *buyerStockPortfolio {
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
			UserID:    buyerID,
			StockID:   stockID,
			StockName: sellerStock.GetStockName(),
			Quantity:  0,
		})

		// Create in database first
		createdStock, err := databaseAccessUser.UserStock().Create(buyerStock)
		if err != nil {
			println("Error: ", err.Error())
			return fmt.Errorf("failed to create buyer stock holding: %v", err)
		}
		buyerStock = createdStock
	}

	// Update Stock quantities in buyer and seller portfolios
	sellerStock.UpdateQuantity(-quantity)
	buyerStock.UpdateQuantity(quantity)

	// Update in database
	err = databaseAccessUser.UserStock().Update(sellerStock)
	if err != nil {
		println("Error: ", err.Error())
		return fmt.Errorf("failed to update seller stock: %v", err)
	}

	err = databaseAccessUser.UserStock().Update(buyerStock)
	if err != nil {
		println("Error: ", err.Error())
		return fmt.Errorf("failed to update buyer stock: %v", err)
	}

	// Update transaction status based on is_buy
	if stockTx.GetIsBuy() {
		// If it's a buy order, set to COMPLETED regardless of partial status
		stockTx.SetOrderStatus("COMPLETED")
	} else {
		// For sell orders, use the existing partial/complete logic
		if !isBuyPartial && !isSellPartial {
			stockTx.SetOrderStatus("COMPLETED")
		} else {
			stockTx.SetOrderStatus("PARTIALLY_COMPLETE")
		}
	}

	// Update the transaction status in database
	err = databaseAccessTransact.StockTransaction().Update(stockTx)
	if err != nil {
		println("Error: ", err.Error())
		return fmt.Errorf("failed to update stock transaction status: %v", err)
	}

	// Create a filled transaction for partial orders only if it's not a buy order
	if (isBuyPartial || isSellPartial) && !stockTx.GetIsBuy() {
		filledTx := transaction.NewStockTransaction(transaction.NewStockTransactionParams{
			NewEntityParams: entity.NewEntityParams{
				DateCreated:  time.Now(),
				DateModified: time.Now(),
			},
			ParentStockTransaction: stockTx,
			OrderStatus:            "COMPLETED",
			TimeStamp:              time.Now(),
		})

		_, err = databaseAccessTransact.StockTransaction().Create(filledTx)
		if err != nil {
			println("Error: ", err.Error())
			return fmt.Errorf("failed to create filled stock transaction: %v", err)
		}
	}

	return nil
}

// Calculates the total cost of a transaction given the quantity and stock price.
func calculateTotalTransactionCost(quantity int, stockPrice float64) float64 {
	return float64(quantity) * stockPrice
}

// Check if buyer has enough funds to afford the quantity*stockprice
// If they dont, return to matching engine that the match was unsuccessful.
func validateBuyerWalletBalance(buyerWallet wallet.WalletInterface, totalCost float64) (bool, error) {
	if buyerWallet == nil {
		return false, errors.New("buyer wallet not found")
	}

	buyerBalance := buyerWallet.GetBalance()

	return buyerBalance >= totalCost, nil
}
