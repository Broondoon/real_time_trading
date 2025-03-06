package orderExecutorService

import (
	//"Shared/entities/entity"
	"Shared/entities/transaction"
	//"Shared/entities/user-stock"
	"Shared/entities/wallet"
	"Shared/network"
	"databaseAccessTransaction"
	"databaseAccessUserManagement"
	"fmt"
	"strings"
)

// ProcessTrade
func ProcessTrade(orderData network.MatchingEngineToExecutionJSON, databaseAccessTransact databaseAccessTransaction.DatabaseAccessInterface, databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface) (bool, bool, error) {

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

	totalCost := calculateTotalTransactionCost(quantity, stockPrice)


	println(fmt.Sprintf(`
	Buyer ID: %s
	Seller ID: %s
	Stock ID: %s
	Buy Order ID: %s
	Sell Order ID: %s
	Is Buy Partial: %t
	Is Sell Partial: %t
	Stock Price: %.2f
	Quantity: %d
	Total Cost: %.2f`, 
		buyerID, 
		sellerID, 
		stockID, 
		buyOrderID, 
		sellOrderID, 
		isBuyPartial, 
		isSellPartial, 
		stockPrice, 
		quantity,
		totalCost))


		
	// 1. Go to the Transaction DB, get stock transactions associated with the buyOrder ID and the sellOrder ID
	transactionList, err := databaseAccessTransact.StockTransaction().GetByIDs([]string{buyOrderID, sellOrderID})
	if err != nil {
		println("Error: ", err.Error())
		return false, false, fmt.Errorf("failed to get transactions: %v", err)
	}
	println("Stock Transactions List:")
	for i, transaction := range *transactionList {
		json, err := transaction.ToJSON()
		if err != nil {
			println("Error converting transaction to JSON: ", err.Error())
			continue // Continue with the next transaction instead of failing
		}
		fmt.Printf("Transaction %d: %s\n", i, string(json))
	}

	// Validate we got both transactions
	if len(*transactionList) != 2 {
		return false, false, fmt.Errorf("expected 2 transactions, got %d", len(*transactionList))
	}

	stockTx := (*transactionList)[0]
	println(fmt.Sprintf("Grabbing first stock transaction in Transaction List: ID= %s, Order Status= %s", stockTx.GetId(), stockTx.GetOrderStatus()))

	// 2. Go to User-Managment DB, get wallet of userID present on  Buy order transaction
	walletList, err := databaseAccessUser.Wallet().GetByIDs([]string{buyerID, sellerID})
	if err != nil {
		println("Error: ", err.Error())
		return false, false, fmt.Errorf("failed to get wallets: %v", err)
	}
	
	println("Wallet List:")
	for i, wallet := range *walletList {
		json, err := wallet.ToJSON()
		if err != nil {
			println("Error converting wallet to JSON: ", err.Error())
			continue // Continue with next wallet instead of failing
		}
		fmt.Printf("Wallet %d (UserID: %s): %s\n", i, wallet.GetUserID(), string(json))
	}


	// Validate we got both wallets
	if len(*walletList) != 2 {
		return false, false, fmt.Errorf("expected 2 wallets, got %d", len(*walletList))
	}

	
	var buyerWallet, sellerWallet wallet.WalletInterface
	for _, w := range *walletList {
		if w.GetUserID() == buyerID {
			buyerWallet = w
		} else if w.GetUserID() == sellerID {
			sellerWallet = w
		}
	}

	
	// 3. Check if buyer has enough funds to afford the quantity*stockprice
	buyerHasFunds, err := validateBuyerWalletBalance(buyerWallet, totalCost)
	println("The buyer has enough funds in their wallet?: ", buyerHasFunds)
	if err != nil {
		println("Error: ", err.Error())
		return false, false, err
	}
	if !buyerHasFunds {
		return false, true, nil
	}


	// Ask Kyle about the (*walletList)[0] and (*walletList)[0]
	// 4. Update buyer and seller wallet balances and create wallet transactions for these changes
	err = updateUserWallets(buyerWallet, sellerWallet, totalCost, stockTx, databaseAccessUser, databaseAccessTransact)
	println("Done updating wallets")
	if len(*walletList) != 2 {
		return false, false, fmt.Errorf("expected 2 wallets, got %d", len(*walletList))
	}
	if err != nil {
		println("Error: ", err.Error())
		return false, false, fmt.Errorf("failed to update wallet balances: %v", err)
	}



	// 5. Update buyer and seller stock portfolios. Deduct the quantity from seller and add to buyer
	println("Updating user stocks...")
	err = updateUserStocks(buyerID, sellerID, stockID, quantity, stockTx, databaseAccessUser, 
		databaseAccessTransact, isBuyPartial, isSellPartial, stockPrice)
		if err != nil {
		// Special case: if the seller doesn't have enough shares, the buy succeeds but sell fails
		if err.Error() == "seller does not have enough shares of stock "+stockID ||  // Support old error format
			strings.Contains(err.Error(), "seller does not have enough shares of stock "+stockID) { // Support new error format
			println("Error:", err.Error())
			return true, false, nil // Buy succeeds, sell fails
		}
		
		// Any other error means the entire transaction failed
		println("Error updating user stocks:", err.Error())
		return false, false, fmt.Errorf("failed to update user stocks: %v", err)
	}

	println("Done updating user stocks")

	println("Done processTrade")
	// 6. Return true to the matching engine to indicate that the trade was successful.
	return true, true, nil

}


// Coordinates the wallet update process
func updateUserWallets(
    buyerWallet wallet.WalletInterface,
    sellerWallet wallet.WalletInterface,
    totalCost float64,
    stockTransaction transaction.StockTransactionInterface,
    databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface,
    databaseAccessTransact databaseAccessTransaction.DatabaseAccessInterface,
) error {


    println(fmt.Sprintf("Initial balances - Buyer: %.2f, Seller: %.2f", buyerWallet.GetBalance(), sellerWallet.GetBalance()))



    // Update buyer's wallet (debit)
    if err := updateWalletBalance(buyerWallet, totalCost, true, databaseAccessUser); err != nil {
        return fmt.Errorf("buyer wallet update failed: %v", err)
    }
	println(fmt.Sprintf("Buyer wallet updated - New balance: %.2f (deducted %.2f)", 
    buyerWallet.GetBalance(), totalCost))


    // Update seller's wallet (credit)
    if err := updateWalletBalance(sellerWallet, totalCost, false, databaseAccessUser); err != nil {
        // Rollback buyer's wallet if seller update fails
        updateWalletBalance(buyerWallet, totalCost, false, databaseAccessUser)
        return fmt.Errorf("seller wallet update failed: %v", err)
    }
	println(fmt.Sprintf("Seller wallet updated - New balance: %.2f (added %.2f)", 
    sellerWallet.GetBalance(), totalCost))




    // Create wallet transactions
    buyerWalletTxID, err := createWalletTransaction(buyerWallet, stockTransaction, true, totalCost, databaseAccessTransact)
    if err != nil {
        return fmt.Errorf("buyer wallet transaction failed: %v", err)
    }
    println(fmt.Sprintf("Created wallet transaction for buyer (ID: %s, UserID: %s) - Amount: %.2f (debit)", 
        buyerWalletTxID, buyerWallet.GetUserID(), totalCost))


	
    sellerWalletTxID, err := createWalletTransaction(sellerWallet, stockTransaction, false, totalCost, databaseAccessTransact)
    if err != nil {
        return fmt.Errorf("seller wallet transaction failed: %v", err)
    }
    println(fmt.Sprintf("Created wallet transaction for seller (ID: %s, UserID: %s) - Amount: %.2f (credit)", 
        sellerWalletTxID, sellerWallet.GetUserID(), totalCost))



    println(fmt.Sprintf("Final balances - Buyer: %.2f, Seller: %.2f",
        buyerWallet.GetBalance(),
        sellerWallet.GetBalance()))


    return nil
}




// Coordinates the stock update process
func updateUserStocks(
    buyerID string,
    sellerID string,
    stockID string,
    quantity int,
	stockTx transaction.StockTransactionInterface,
    databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface,
    databaseAccessTransact databaseAccessTransaction.DatabaseAccessInterface,
    isBuyPartial bool,
    isSellPartial bool,
	stockPrice float64,
) error {


    buyerPortfolio, sellerPortfolio, err := findUserStockPortfolios(buyerID, sellerID, databaseAccessUser)
    if err != nil {
        return err
    }
	println(fmt.Sprintf("Step 1: Successfully retrieved user stock portfolios - Buyer: %s (%d stocks), Seller: %s (%d stocks)",
	buyerID, len(*buyerPortfolio), sellerID, len(*sellerPortfolio)))


    sellerStock, err := handleSellerStock(sellerPortfolio, stockID, quantity)
    if err != nil {
        return err
    }
	println(fmt.Sprintf("Step 2: Successfully validated seller's stock - Seller has %d shares of %s", 
	sellerStock.GetQuantity(), stockID))


    buyerStock, err := handleBuyerStock(buyerPortfolio, buyerID, stockID, sellerStock, databaseAccessUser)
    if err != nil {
        return err
    }
    println(fmt.Sprintf("Step 3: Successfully retrieved/created buyer's stock - Buyer initially has %d shares of %s", 
        buyerStock.GetQuantity(), stockID))

    if err := updateUserStockQuantities(buyerStock, sellerStock, quantity, databaseAccessUser); err != nil {
        return err
    }
    println(fmt.Sprintf("Step 4: Successfully updated stock quantities - Transferred %d shares from seller to buyer", 
        quantity))


    if err := updateTransactionStatus(stockTx, isBuyPartial, isSellPartial, stockPrice, databaseAccessTransact); err != nil {
        return err
    }
    println(fmt.Sprintf("Step 5: Successfully updated transaction status - Buy Partial: %t, Sell Partial: %t", 
        isBuyPartial, isSellPartial))


	println("All stock operations completed successfully")
    return nil
}




