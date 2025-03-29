package orderExecutorService

// import (
// 	"Shared/entities/transaction"
// 	"Shared/entities/wallet"
// 	"databaseAccessTransaction"
// 	"databaseAccessUserManagement"
// 	"fmt"
// )

// // Check if buyer has enough funds to afford the quantity*stockprice
// // If they don't, return to matching engine that the match was unsuccessful.
// // func validateBuyerWalletBalance(buyerWallet wallet.WalletInterface, totalCost float64) (bool, error) {
// // 	if buyerWallet == nil {
// // 		return false, errors.New("buyer wallet not found")
// // 	}

// // 	buyerBalance := buyerWallet.GetBalance()

// // 	return buyerBalance >= totalCost, nil
// // }

// // Updates the balance of a single wallet and handles errors
// func updateWalletBalance(
// 	buyerWallet wallet.WalletInterface,
// 	sellerWallet wallet.WalletInterface,
// 	amount float64,
// 	databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface,
// ) error {
// 	buyerWallet.UpdateBalance(-amount)
// 	sellerWallet.UpdateBalance(amount)

// 	mapErr, err := databaseAccessUser.Wallet().UpdateBulk(&[]wallet.WalletInterface{buyerWallet, sellerWallet})
// 	if err != nil || len(mapErr) > 0 {
// 		// Rollback the balance change if update fails
// 		sellerWallet.UpdateBalance(-amount)
// 		buyerWallet.UpdateBalance(amount)
// 		return fmt.Errorf("failed to update wallet balance: %v", err)
// 	}
// 	return nil
// }

// // Creates a wallet transaction record and returns its ID
// func createWalletTransaction(
// 	buyerWallet wallet.WalletInterface,
// 	buyerStockTransaction transaction.StockTransactionInterface,
// 	sellerWallet wallet.WalletInterface,
// 	sellerStockTransaction transaction.StockTransactionInterface,
// 	amount float64,
// 	databaseAccessTransact databaseAccessTransaction.DatabaseAccessInterface,
// ) (transaction.WalletTransactionInterface, transaction.WalletTransactionInterface, error) {
// 	buyerWalletTx := transaction.NewWalletTransaction(transaction.NewWalletTransactionParams{
// 		NewTransactionParams: &transaction.NewTransactionParams{
// 			UserID: buyerWallet.GetUserID(),
// 		},
// 		WalletID:           buyerWallet.GetId(),
// 		StockTransactionID: buyerStockTransaction.GetId(),
// 		IsDebit:            true,
// 		Amount:             amount,
// 		Wallet:             buyerWallet,
// 		StockTransaction:   buyerStockTransaction,
// 	})
// 	sellerWalletTx := transaction.NewWalletTransaction(transaction.NewWalletTransactionParams{
// 		NewTransactionParams: &transaction.NewTransactionParams{
// 			UserID: sellerWallet.GetUserID(),
// 		},
// 		WalletID:           sellerWallet.GetId(),
// 		StockTransactionID: sellerStockTransaction.GetId(),
// 		IsDebit:            false,
// 		Amount:             amount,
// 		Wallet:             sellerWallet,
// 		StockTransaction:   sellerStockTransaction,
// 	})

// 	createdTxs, errMap, err := databaseAccessTransact.WalletTransaction().CreateBulk(&[]transaction.WalletTransactionInterface{buyerWalletTx, sellerWalletTx})
// 	if err != nil || len(errMap) > 0 {
// 		return nil, nil, fmt.Errorf("failed to create wallet transaction: %v", err)
// 	}

// 	for _, tx := range *createdTxs {
// 		if tx.GetWalletIDString() == buyerWallet.GetIdString() {
// 			buyerWalletTx = tx.(*transaction.WalletTransaction)
// 		} else if tx.GetWalletIDString() == sellerWallet.GetIdString() {
// 			sellerWalletTx = tx.(*transaction.WalletTransaction)
// 		}
// 	}
// 	return buyerWalletTx, sellerWalletTx, nil
// }
