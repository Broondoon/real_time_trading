package orderExecutorService

import (
    "Shared/entities/entity"
    "Shared/entities/transaction"
    "Shared/entities/wallet"
    "databaseAccessTransaction"
    "databaseAccessUserManagement"
    "fmt"
    "time"
	"errors"
)



// Check if buyer has enough funds to afford the quantity*stockprice
// If they dont, return to matching engine that the match was unsuccessful.
func validateBuyerWalletBalance(buyerWallet wallet.WalletInterface, totalCost float64) (bool, error) {
    if buyerWallet == nil {
        return false, errors.New("buyer wallet not found")
    }

    buyerBalance := buyerWallet.GetBalance()

    return buyerBalance >= totalCost, nil
}





// Updates the balance of a single wallet and handles errors
func updateWalletBalance(
    wallet wallet.WalletInterface,
    amount float64,
    isDebit bool,
    databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface,
) error {
    initialBalance := wallet.GetBalance()
    if isDebit {
        wallet.SetBalance(initialBalance - amount)
    } else {
        wallet.SetBalance(initialBalance + amount)
    }

    err := databaseAccessUser.Wallet().Update(wallet)
    if err != nil {
        // Rollback the balance change if update fails
        wallet.SetBalance(initialBalance)
        return fmt.Errorf("failed to update wallet balance: %v", err)
    }
    return nil
}





// Creates a wallet transaction record and returns its ID
func createWalletTransaction(
    wallet wallet.WalletInterface,
    stockTransaction transaction.StockTransactionInterface,
    isDebit bool,
    amount float64,
    databaseAccessTransact databaseAccessTransaction.DatabaseAccessInterface,
) (string, error) {
    walletTx := transaction.NewWalletTransaction(transaction.NewWalletTransactionParams{
        NewEntityParams: entity.NewEntityParams{
            DateCreated:  time.Now(),
            DateModified: time.Now(),
        },
        WalletID:           wallet.GetId(),
        StockTransactionID: stockTransaction.GetId(),
        IsDebit:            isDebit,
        Amount:             amount,
        Timestamp:          time.Now(),
        Wallet:             wallet,
        StockTransaction:   stockTransaction,
        UserID:             wallet.GetUserID(),
    })

    createdTx, err := databaseAccessTransact.WalletTransaction().Create(walletTx)
    if err != nil {
        return "", fmt.Errorf("failed to create wallet transaction: %v", err)
    }
    
    // Set wallet transaction ID and return it
    createdTx.SetWalletTXID()
    return createdTx.GetId(), nil
}
