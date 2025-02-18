package databaseServiceStockTransactions

import (
	databaseService "Shared/database/database-service"
	"Shared/entities/transaction"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DatabaseServiceInterface interface {
	databaseService.PostGresDatabaseInterface
	GetStockTransaction(transactionID string) (transaction.StockTransactionInterface, error)
	GetStockTransactions(transactionIDs *[]string) (*[]transaction.StockTransactionInterface, error)
	//GetInitialStockTransactionsForStock(stockID string) (*[]transaction.StockTransactionInterface, error)
	CreateStockTransaction(StockTransaction transaction.StockTransactionInterface) (transaction.StockTransactionInterface, error)
	UpdateStockTransaction(StockTransaction transaction.StockTransactionInterface) (transaction.StockTransactionInterface, error)
	DeleteStockTransaction(transactionID string) (transaction.StockTransactionInterface, error)
}

type DatabaseService struct {
	databaseService.PostGresDatabaseInterface
	StockTransactions *[]transaction.StockTransactionInterface
}

type NewDatabaseServiceParams struct {
	*databaseService.NewPostGresDatabaseParams
}

func NewDatabaseService(params NewDatabaseServiceParams) DatabaseServiceInterface {
	db := &DatabaseService{
		PostGresDatabaseInterface: databaseService.NewPostGresDatabase(params.NewPostGresDatabaseParams),
	}
	db.Connect()
	db.GetDatabaseConnection().AutoMigrate(&transaction.StockTransaction{})
	return db
}

func (d *DatabaseService) GetStockTransaction(transactionID string) (transaction.StockTransactionInterface, error) {
	var transaction transaction.StockTransaction
	d.GetDatabaseConnection().Where("ID = ?", transactionID).First(&transaction)
	if errors.Is(d.GetDatabaseConnection().Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("stock transaction not found: %s", transactionID)
	}
	transaction.SetDefaults()
	return &transaction, nil

}

func (d *DatabaseService) GetStockTransactions(transactionIDs *[]string) (*[]transaction.StockTransactionInterface, error) {
	var transactions []transaction.StockTransaction
	d.GetDatabaseConnection().Where("ID IN ?", transactionIDs).Find(&transactions)
	if errors.Is(d.GetDatabaseConnection().Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("stock transactions not found: %v", transactionIDs)
	}
	for _, o := range transactions {
		o.SetDefaults()
	}
	converted := make([]transaction.StockTransactionInterface, len(transactions))
	for i := range transactions {
		converted[i] = &(transactions)[i]
	}
	return &converted, nil
}

// // Right now, we're just gonna get all stockstransactions for a given stock. Later, we need to limit this to a specific subset of transactions.
// func (d *DatabaseService) GetInitialStockTransactionsForStock(stockID string) (*[]transaction.StockTransactionInterface, error) {
// 	var transactions []transaction.StockTransaction
// 	d.GetDatabaseConnection().Where("StockID = ? ", stockID).Find(&transactions)
// 	if errors.Is(d.GetDatabaseConnection().Error, gorm.ErrRecordNotFound) {
// 		return nil, fmt.Errorf("stock transactions not found")
// 	}
// 	for _, o := range transactions {
// 		o.SetDefaults()
// 	}
// 	converted := make([]transaction.StockTransactionInterface, len(transactions))
// 	for i := range transactions {
// 		converted[i] = &(transactions)[i]
// 	}
// 	return &converted, nil
// }

// Generated with CHATGPT Because copilot was just screwing around with me. All hail the mighty AI.
func (d *DatabaseService) CreateStockTransaction(StockTransaction transaction.StockTransactionInterface) (transaction.StockTransactionInterface, error) {
	// Generate a unique transaction ID. Because a UUID collision is extraordinarily unlikely,
	// most systems skip the existence check altogether.
	// However, if you REALLY want to confirm the ID doesn't already exist
	// (for example, for short random strings, not standard UUIDs), you can do so in a loop.
	var existing transaction.StockTransaction
	for {
		candidateID := generateRandomtransactionID()

		// Check if it already exists in the database
		result := d.GetDatabaseConnection().
			Where("ID = ?", candidateID).
			First(&existing)

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			StockTransaction.SetId(candidateID)
			break
		}

		// If there's another error, handle or return it
		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("error checking existing transaction: %w", result.Error)
		}
		// If record already found, loop again and generate a new ID
		// (though collisions for standard UUIDs are extremely unlikely)
	}

	// Once we have a unique ID assigned, create the record
	createResult := d.GetDatabaseConnection().Create(StockTransaction.(*transaction.StockTransaction))
	if createResult.Error != nil {
		return nil, fmt.Errorf("error creating stock transaction: %w", createResult.Error)
	}

	return StockTransaction, nil
}

func generateRandomtransactionID() string {
	// Generate a new UUID as the transaction ID
	return uuid.New().String()
}

func (d *DatabaseService) UpdateStockTransaction(StockTransaction transaction.StockTransactionInterface) (transaction.StockTransactionInterface, error) {
	// Update the record in the database
	updateResult := d.GetDatabaseConnection().Save(StockTransaction.(*transaction.StockTransaction))
	if updateResult.Error != nil {
		return nil, fmt.Errorf("error updating stock transaction: %w", updateResult.Error)
	}

	return StockTransaction, nil
}

func (d *DatabaseService) DeleteStockTransaction(transactionID string) (transaction.StockTransactionInterface, error) {
	// Get the transaction to delete
	StockTransaction, err := d.GetStockTransaction(transactionID)
	if err != nil {
		return nil, fmt.Errorf("error getting stock transaction: %w", err)
	}

	// Delete the record from the database
	deleteResult := d.GetDatabaseConnection().Delete(StockTransaction.(*transaction.StockTransaction), transactionID)
	if deleteResult.Error != nil {
		return nil, fmt.Errorf("error deleting stock transaction: %w", deleteResult.Error)
	}

	return StockTransaction, nil
}
