package databaseServiceTransaction

import (
	databaseService "Shared/database/database-service"
	"Shared/entities/transaction"
	"os"
	"time"

	"gorm.io/gorm"
)

type StockTransactionDataServiceInterface = databaseService.EntityDataInterface[*transaction.StockTransaction, *gorm.DB]
type WalletTransactionDataServiceInterface = databaseService.EntityDataInterface[*transaction.WalletTransaction, *gorm.DB]

type DatabaseServiceInterface interface {
	databaseService.DatabaseInterface[*gorm.DB]
	StockTransactions() StockTransactionDataServiceInterface
	WalletTransactions() WalletTransactionDataServiceInterface
}

type DatabaseService struct {
	StockTransaction  StockTransactionDataServiceInterface
	WalletTransaction WalletTransactionDataServiceInterface
	databaseService.DatabaseInterface[*gorm.DB]
}

type NewDatabaseServiceParams struct {
	// StockTransactionParams  *databaseService.NewEntityDataParams // leave nil for default
	// WalletTransactionParams *databaseService.NewEntityDataParams // leave nil for default
	//Only the StockTransactionParams.NewPostGresDatabaseParams is used. The WalletTransactionParams.NewPostGresDatabaseParams is ignored.
}

func NewDatabaseService(params *NewDatabaseServiceParams) DatabaseServiceInterface {
	// if params.StockTransactionParams == nil {
	// 	params.StockTransactionParams = &databaseService.NewEntityDataParams{
	// 		NewPostGresDatabaseParams: &databaseService.NewPostGresDatabaseParams{},
	// 	}
	// }
	// if params.WalletTransactionParams == nil {
	// 	params.WalletTransactionParams = &databaseService.NewEntityDataParams{
	// 		NewPostGresDatabaseParams: &databaseService.NewPostGresDatabaseParams{},
	// 	}
	// }
	newDBConnection := databaseService.NewPostGresDatabase(&databaseService.NewPostGresDatabaseParams{})
	// if params.StockTransactionParams.Existing != nil {
	// 	newDBConnection = params.StockTransactionParams.Existing
	// 	if params.WalletTransactionParams.Existing == nil {
	// 		params.WalletTransactionParams.Existing = newDBConnection
	// 	}
	// } else if params.WalletTransactionParams.Existing != nil {
	// 	newDBConnection = params.WalletTransactionParams.Existing
	// 	params.StockTransactionParams.Existing = newDBConnection
	// } else {
	// 	newDBConnection = databaseService.NewPostGresDatabase(params.StockTransactionParams.NewPostGresDatabaseParams)
	// 	params.StockTransactionParams.Existing = newDBConnection
	// 	params.WalletTransactionParams.Existing = newDBConnection
	// }

	//CACHE IMPLEMENTATION
	/* cachedStockTransaction := databaseService.NewCachedEntityData[*transaction.StockTransaction](&databaseService.NewCachedEntityDataParams{
		NewEntityDataParams: params.StockTransactionParams,
		RedisAddr:           os.Getenv("REDIS_ADDR"),
		Password:            os.Getenv("REDIS_PASSWORD"),
		DefaultTTL:          5 * time.Minute,
	})

	cachedWalletTransaction := databaseService.NewCachedEntityData[*transaction.WalletTransaction](&databaseService.NewCachedEntityDataParams{
		NewEntityDataParams: params.WalletTransactionParams,
		RedisAddr:           os.Getenv("REDIS_ADDR"),
		Password:            os.Getenv("REDIS_PASSWORD"),
		DefaultTTL:          5 * time.Minute,
	})

	db := &DatabaseService{
		StockTransaction:  cachedStockTransaction,
		WalletTransaction: cachedWalletTransaction,
		DatabaseInterface: newDBConnection,
	} */

	db := &DatabaseService{
		StockTransaction: databaseService.NewCachedEntityData[*transaction.StockTransaction, *gorm.DB](&databaseService.NewCachedEntityDataParams[*transaction.StockTransaction, *gorm.DB]{
			RedisAddr:  os.Getenv("REDIS_ADDR"),
			Password:   os.Getenv("REDIS_PASSWORD"),
			DefaultTTL: 5 * time.Minute,
			EntityData: databaseService.NewPostGresEntityData[*transaction.StockTransaction](
				&databaseService.NewPostGresEntityDataParams{
					Existing: newDBConnection,
				},
			),
		}),
		WalletTransaction: databaseService.NewCachedEntityData[*transaction.WalletTransaction, *gorm.DB](&databaseService.NewCachedEntityDataParams[*transaction.WalletTransaction, *gorm.DB]{
			RedisAddr:  os.Getenv("REDIS_ADDR"),
			Password:   os.Getenv("REDIS_PASSWORD"),
			DefaultTTL: 5 * time.Minute,
			EntityData: databaseService.NewPostGresEntityData[*transaction.WalletTransaction](
				&databaseService.NewPostGresEntityDataParams{
					Existing: newDBConnection,
				},
			),
		}),
		DatabaseInterface: newDBConnection,
	}
	db.Connect()
	db.StockTransactions().GetDatabaseSession().AutoMigrate(&transaction.StockTransaction{})
	db.WalletTransactions().GetDatabaseSession().AutoMigrate(&transaction.WalletTransaction{})
	return db
}

func (d *DatabaseService) StockTransactions() StockTransactionDataServiceInterface {
	return d.StockTransaction
}

func (d *DatabaseService) WalletTransactions() WalletTransactionDataServiceInterface {
	return d.WalletTransaction
}

func (d *DatabaseService) Connect() {
	d.StockTransactions().Connect()
	d.WalletTransactions().Connect()
}

func (d *DatabaseService) Disconnect() {
	d.StockTransactions().Disconnect()
	d.WalletTransactions().Disconnect()
}
