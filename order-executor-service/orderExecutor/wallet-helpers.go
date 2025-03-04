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





// Creates a wallet transaction record
func createWalletTransaction(
    wallet wallet.WalletInterface,
    stockTransaction transaction.StockTransactionInterface,
    isDebit bool,
    amount float64,
    databaseAccessTransact databaseAccessTransaction.DatabaseAccessInterface,
) error {
    walletTx := transaction.NewWalletTransaction(transaction.NewWalletTransactionParams{
        NewEntityParams: entity.NewEntityParams{
            DateCreated:  time.Now(),
            DateModified: time.Now(),
        },
        Wallet:           wallet,
        StockTransaction: stockTransaction,
        IsDebit:          isDebit,
        Amount:           amount,
        Timestamp:        time.Now(),
    })

    _, err := databaseAccessTransact.WalletTransaction().Create(walletTx)
    if err != nil {
        return fmt.Errorf("failed to create wallet transaction: %v", err)
    }
    return nil
}
