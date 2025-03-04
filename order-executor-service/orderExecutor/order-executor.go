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


	// Validate we got both transactions
	if len(*transactionList) != 2 {
		return false, false, fmt.Errorf("expected 2 transactions, got %d", len(*transactionList))
	}

	stockTx := (*transactionList)[0]

	var buyTx, sellTx transaction.StockTransactionInterface
	for _, tx := range *transactionList {
		if tx.GetId() == buyOrderID {
			buyTx = tx
		} else if tx.GetId() == sellOrderID {
			sellTx = tx
		}
	}

	if buyTx == nil || sellTx == nil {
		return false, false, fmt.Errorf("could not match buy and sell transactions")
	}



	// 2. Go to User-Managment DB, get wallet of userID present on  Buy order transaction
	walletList, err := databaseAccessUser.Wallet().GetByIDs([]string{buyerID, sellerID})
	if err != nil {
		println("Error: ", err.Error())
		return false, false, fmt.Errorf("failed to get wallets: %v", err)
	}

	// Validate we got both wallets
	if len(*walletList) != 2 {
		return false, false, fmt.Errorf("expected 2 wallets, got %d", len(*walletList))
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



	// Ask Kyle about the (*walletList)[0] and (*walletList)[0]
	// 4. Update buyer and seller wallet balances and create wallet transactions for these changes
	err = updateUserWallets((*walletList)[0], (*walletList)[1], totalCost, stockTx, databaseAccessUser, databaseAccessTransact)
	if len(*walletList) != 2 {
		return false, false, fmt.Errorf("expected 2 wallets, got %d", len(*walletList))
	}
	if err != nil {
		println("Error: ", err.Error())
		return false, false, fmt.Errorf("failed to update wallet balances: %v", err)
	}



	// 5. Update buyer and seller stock portfolios. Deduct the stock quantity from the seller's portfolio.
	//    Add the stock quantity to the buyer's portfolio.
	err = updateUserStocks(buyerID, sellerID, stockID, quantity, buyTx, sellTx, databaseAccessUser, databaseAccessTransact, isBuyPartial, isSellPartial)
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




// Coordinates the wallet update process
func updateUserWallets(
    buyerWallet wallet.WalletInterface,
    sellerWallet wallet.WalletInterface,
    totalCost float64,
    stockTransaction transaction.StockTransactionInterface,
    databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface,
    databaseAccessTransact databaseAccessTransaction.DatabaseAccessInterface,
) error {


    println("Initial balances - Buyer: %.2f, Seller: %.2f", buyerWallet.GetBalance(), sellerWallet.GetBalance())



    // Update buyer's wallet (debit)
    if err := updateWalletBalance(buyerWallet, totalCost, true, databaseAccessUser); err != nil {
        return fmt.Errorf("buyer wallet update failed: %v", err)
    }



    // Update seller's wallet (credit)
    if err := updateWalletBalance(sellerWallet, totalCost, false, databaseAccessUser); err != nil {
        // Rollback buyer's wallet if seller update fails
        updateWalletBalance(buyerWallet, totalCost, false, databaseAccessUser)
        return fmt.Errorf("seller wallet update failed: %v", err)
    }



    // Create wallet transactions
    if err := createWalletTransaction(buyerWallet, stockTransaction, true, totalCost, databaseAccessTransact); err != nil {
        return fmt.Errorf("buyer wallet transaction failed: %v", err)
    }

    if err := createWalletTransaction(sellerWallet, stockTransaction, false, totalCost, databaseAccessTransact); err != nil {
        return fmt.Errorf("seller wallet transaction failed: %v", err)
    }



    println("Final balances - Buyer: %.2f, Seller: %.2f",
        buyerWallet.GetBalance(),
        sellerWallet.GetBalance())

    return nil

}



// Coordinates the stock update process
func updateUserStocks(
    buyerID string,
    sellerID string,
    stockID string,
    quantity int,
    buyTx transaction.StockTransactionInterface,
	sellTx transaction.StockTransactionInterface,
    databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface,
    databaseAccessTransact databaseAccessTransaction.DatabaseAccessInterface,
    isBuyPartial bool,
    isSellPartial bool,
) error {


    buyerPortfolio, sellerPortfolio, err := getUserStockPortfolios(buyerID, sellerID, databaseAccessUser)
    if err != nil {
        return err
    }


    sellerStock, err := handleSellerStock(sellerPortfolio, stockID, quantity)
    if err != nil {
        return err
    }


    buyerStock, err := handleBuyerStock(buyerPortfolio, buyerID, stockID, sellerStock, databaseAccessUser)
    if err != nil {
        return err
    }


    if err := updateUserStockQuantities(buyerStock, sellerStock, quantity, databaseAccessUser); err != nil {
        return err
    }



    if err := updateTransactionStatus(buyTx, sellTx, isBuyPartial, isSellPartial, databaseAccessTransact); err != nil {
        return err
    }



    return nil
}




