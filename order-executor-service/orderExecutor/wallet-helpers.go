package orderExecutorService

import (
	"Shared/entities/entity"
	"Shared/entities/transaction"
	"Shared/entities/wallet"
	"databaseAccessTransaction"
	"databaseAccessUserManagement"
	"errors"
	"fmt"
	"time"
)

// Check if buyer has enough funds to afford the quantity*stockprice
// If they don't, return to matching engine that the match was unsuccessful.
func validateBuyerWalletBalance(buyerWallet wallet.WalletInterface, totalCost float64) (bool, error) {
	if buyerWallet == nil {
		return false, errors.New("buyer wallet not found")
	}

	buyerBalance := buyerWallet.GetBalance()

	return buyerBalance >= totalCost, nil
}

// Updates the balance of a single wallet and handles errors
func updateWalletBalance(
	buyerWallet wallet.WalletInterface,
	sellerWallet wallet.WalletInterface,
	amount float64,
	databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface,
) error {
	buyerWallet.UpdateBalance(-amount)
	sellerWallet.UpdateBalance(amount)

	mapErr, err := databaseAccessUser.Wallet().UpdateBulk(&[]wallet.WalletInterface{buyerWallet, sellerWallet})
	if err != nil || len(mapErr) > 0 {
		// Rollback the balance change if update fails
		sellerWallet.UpdateBalance(-amount)
		buyerWallet.UpdateBalance(amount)
		return fmt.Errorf("failed to update wallet balance: %v", err)
	}
	return nil
}

// Creates a wallet transaction record and returns its ID
func createWalletTransaction(
	buyerWallet wallet.WalletInterface,
	buyerStockTransaction transaction.StockTransactionInterface,
	sellerWallet wallet.WalletInterface,
	sellerStockTransaction transaction.StockTransactionInterface,
	amount float64,
	databaseAccessTransact databaseAccessTransaction.DatabaseAccessInterface,
) (string, string, error) {
	buyerWalletTx := transaction.NewWalletTransaction(transaction.NewWalletTransactionParams{
		NewEntityParams: entity.NewEntityParams{
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		WalletID:           buyerWallet.GetId(),
		StockTransactionID: buyerStockTransaction.GetId(),
		IsDebit:            true,
		Amount:             amount,
		Timestamp:          time.Now(),
		Wallet:             buyerWallet,
		StockTransaction:   buyerStockTransaction,
		UserID:             buyerWallet.GetUserID(),
	})
	sellerWalletTx := transaction.NewWalletTransaction(transaction.NewWalletTransactionParams{
		NewEntityParams: entity.NewEntityParams{
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		WalletID:           sellerWallet.GetId(),
		StockTransactionID: sellerStockTransaction.GetId(),
		IsDebit:            false,
		Amount:             amount,
		Timestamp:          time.Now(),
		Wallet:             sellerWallet,
		StockTransaction:   sellerStockTransaction,
		UserID:             sellerWallet.GetUserID(),
	})

	createdTxs, errMap, err := databaseAccessTransact.WalletTransaction().CreateBulk(&[]transaction.WalletTransactionInterface{buyerWalletTx, sellerWalletTx})
	if err != nil || len(errMap) > 0 {
		return "", "", fmt.Errorf("failed to create wallet transaction: %v", err)
	}

	var buyerWalletTxID string
	var sellerWalletTxID string
	for _, tx := range *createdTxs {
		if tx.GetWalletIDString() == buyerWallet.GetIdString() {
			buyerWalletTxID = tx.GetIdString()
		} else if tx.GetWalletIDString() == sellerWallet.GetIdString() {
			sellerWalletTxID = tx.GetIdString()
		}
	}
	return buyerWalletTxID, sellerWalletTxID, nil
}
