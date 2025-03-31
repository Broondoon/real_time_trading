package databaseServiceUserManagement

import (
	databaseService "Shared/database/database-service"
	"Shared/entities/transaction"
	userStock "Shared/entities/user-stock"
	"Shared/entities/wallet"
	"Shared/network"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StockTransactionDataServiceInterface = databaseService.EntityDataInterface[*transaction.StockTransaction, *gorm.DB]
type WalletTransactionDataServiceInterface = databaseService.EntityDataInterface[*transaction.WalletTransaction, *gorm.DB]

type UserStockDataServiceInterface interface {
	databaseService.EntityDataInterface[*userStock.UserStock, *gorm.DB]
}

type WalletDataServiceInterface interface {
	databaseService.EntityDataInterface[*wallet.Wallet, *gorm.DB]
}

type DatabaseServiceInterface interface {
	databaseService.DatabaseInterface[*gorm.DB]
	UserStocks() UserStockDataServiceInterface
	Wallets() WalletDataServiceInterface
	StockTransactions() StockTransactionDataServiceInterface
	WalletTransactions() WalletTransactionDataServiceInterface
	ExecuteOrder(buyerStockID string, sellerStockID string, buyerID string, sellerID string, stockID string, stockName string, stockPrice float64, quantity int) (network.ExecutorToMatchingEngineJSON, error)
}

type DatabaseService struct {
	UserStock         databaseService.EntityDataInterface[*userStock.UserStock, *gorm.DB]
	Wallet            databaseService.EntityDataInterface[*wallet.Wallet, *gorm.DB]
	StockTransaction  StockTransactionDataServiceInterface
	WalletTransaction WalletTransactionDataServiceInterface
	databaseService.DatabaseInterface[*gorm.DB]
}

type NewDatabaseServiceParams struct {
	// UserStockParams *databaseService.NewEntityDataParams // leave nil for default
	// WalletParams    *databaseService.NewEntityDataParams // leave nil for default
	// StockTransactionParams  *databaseService.NewEntityDataParams // leave nil for default
	// WalletTransactionParams *databaseService.NewEntityDataParams // leave nil for default
	// Only the UserStockParams.NewPostGresDatabaseParams is used. The WalletParams.NewPostGresDatabaseParams is ignored.
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
	// if params.UserStockParams == nil {
	// 	params.UserStockParams = &databaseService.NewEntityDataParams{
	// 		NewPostGresDatabaseParams: &databaseService.NewPostGresDatabaseParams{},
	// 	}
	// }
	// if params.WalletParams == nil {
	// 	params.WalletParams = &databaseService.NewEntityDataParams{
	// 		NewPostGresDatabaseParams: &databaseService.NewPostGresDatabaseParams{},
	// 	}
	// }

	// var newDBConnection databaseService.PostGresDatabaseInterface
	// if params.UserStockParams.Existing != nil {
	// 	newDBConnection = params.UserStockParams.Existing
	// 	if params.WalletParams.Existing == nil {
	// 		params.WalletParams.Existing = newDBConnection
	// 	}
	// } else if params.WalletParams.Existing != nil {
	// 	newDBConnection = params.WalletParams.Existing
	// 	params.UserStockParams.Existing = newDBConnection
	// } else {
	// 	newDBConnection = databaseService.NewPostGresDatabase(params.UserStockParams.NewPostGresDatabaseParams)
	// 	params.UserStockParams.Existing = newDBConnection
	// 	params.WalletParams.Existing = newDBConnection
	// }

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

	//Cache stuff
	/* cachedUserStock := databaseService.NewCachedEntityData[*userStock.UserStock](&databaseService.NewCachedEntityDataParams{
		NewEntityDataParams: params.UserStockParams,
		RedisAddr:           os.Getenv("REDIS_ADDR"),
		Password:            os.Getenv("REDIS_PASSWORD"),
		DefaultTTL:          5 * time.Minute,
	})

	cachedWallet := databaseService.NewCachedEntityData[*wallet.Wallet](&databaseService.NewCachedEntityDataParams{
		NewEntityDataParams: params.WalletParams,
		RedisAddr:           os.Getenv("REDIS_ADDR"),
		Password:            os.Getenv("REDIS_PASSWORD"),
		DefaultTTL:          5 * time.Minute,
	})

	db := &DatabaseService{
		UserStock:         cachedUserStock,
		Wallet:            cachedWallet,
		DatabaseInterface: newDBConnection,
	} */

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

	newDBConnection := databaseService.NewPostGresDatabase(&databaseService.NewPostGresDatabaseParams{})

	db := &DatabaseService{
		UserStock: databaseService.NewPostGresEntityData[*userStock.UserStock](&databaseService.NewPostGresEntityDataParams{
			Existing: newDBConnection,
		}),
		Wallet: databaseService.NewPostGresEntityData[*wallet.Wallet](&databaseService.NewPostGresEntityDataParams{
			Existing: newDBConnection,
		}),
		StockTransaction: databaseService.NewPostGresEntityData[*transaction.StockTransaction](&databaseService.NewPostGresEntityDataParams{
			Existing: newDBConnection,
		}),
		WalletTransaction: databaseService.NewPostGresEntityData[*transaction.WalletTransaction](&databaseService.NewPostGresEntityDataParams{
			Existing: newDBConnection,
		}),
		DatabaseInterface: newDBConnection,
	}

	db.Connect()

	db.UserStocks().GetDatabaseSession().AutoMigrate(&userStock.UserStock{})
	db.Wallets().GetDatabaseSession().AutoMigrate(&wallet.Wallet{})
	db.StockTransactions().GetDatabaseSession().AutoMigrate(&transaction.StockTransaction{})
	db.WalletTransactions().GetDatabaseSession().AutoMigrate(&transaction.WalletTransaction{})

	return db
}

func (d *DatabaseService) UserStocks() UserStockDataServiceInterface {
	return d.UserStock
}

func (d *DatabaseService) Wallets() WalletDataServiceInterface {
	return d.Wallet
}

func (d *DatabaseService) StockTransactions() StockTransactionDataServiceInterface {
	return d.StockTransaction
}

func (d *DatabaseService) WalletTransactions() WalletTransactionDataServiceInterface {
	return d.WalletTransaction
}

func (d *DatabaseService) Connect() {
	d.UserStocks().Connect()
	d.Wallets().Connect()
	d.StockTransactions().Connect()
	d.WalletTransactions().Connect()
}

func (d *DatabaseService) Disconnect() {
	d.UserStocks().Disconnect()
	d.Wallets().Disconnect()
	d.StockTransactions().Disconnect()
	d.WalletTransactions().Disconnect()
}

func (d *DatabaseService) ExecuteOrder(buyerStockID string, sellerStockID string, buyerID string, sellerID string, stockID string, stockName string, stockPrice float64, quantity int) (network.ExecutorToMatchingEngineJSON, error) {
	log.Println("Service: ExecuteOrder")
	log.Println("Buyer Stock ID: ", buyerStockID)
	log.Println("Seller Stock ID: ", sellerStockID)
	log.Println("Buyer ID: ", buyerID)
	log.Println("Seller ID: ", sellerID)
	log.Println("Stock ID: ", stockID)
	log.Println("Stock Name: ", stockName)
	log.Println("Stock Price: ", stockPrice)
	log.Println("Quantity: ", quantity)

	returnStuct := network.ExecutorToMatchingEngineJSON{
		// StockID:            stockID,
		// BuyerStockOrderId:  buyerStockID,
		// SellerStockOrderId: sellerStockID,
		IsBuyFailure:  false,
		IsSellFailure: false,
	}

	buyerStockUUID, err := uuid.Parse(buyerStockID)
	if err != nil {
		returnStuct.IsBuyFailure = true
		return returnStuct, err
	}

	sellerStockUUID, err := uuid.Parse(sellerStockID)
	if err != nil {
		returnStuct.IsSellFailure = true
		return returnStuct, err
	}

	buyerUUID, err := uuid.Parse(buyerID)
	if err != nil {
		returnStuct.IsBuyFailure = true
		return returnStuct, err
	}

	sellerUUID, err := uuid.Parse(sellerID)
	if err != nil {
		returnStuct.IsSellFailure = true
		return returnStuct, err
	}

	stockUUID, err := uuid.Parse(stockID)
	if err != nil {
		returnStuct.IsBuyFailure = true
		returnStuct.IsSellFailure = true
		return returnStuct, err
	}
	timeStamp := time.Now()
	//EXPLAIN (ANALYZE, BUFFERS)
	err = d.GetDatabaseSession().Exec(`
    SELECT process_stock_trade(?, ?, ?, ?, ?, ?, ?, ?);`,
		buyerStockUUID,
		sellerStockUUID,
		buyerUUID,
		sellerUUID,
		stockUUID,
		stockPrice,
		quantity,
		stockName).Error
	if err != nil {
		println("Error in ExecuteOrder: ", err.Error())
		switch err.Error() {
		case `No wallet found for user_id = ` + buyerID:
			returnStuct.IsBuyFailure = true
		case `'balance (user_id=` + buyerID + `) would go negative. Rolling back.`:
			returnStuct.IsBuyFailure = true
		case `'No stock_transaction found for tx_id = ` + buyerStockID:
			returnStuct.IsBuyFailure = true
		case `No wallet found for user_id = ` + sellerID:
			returnStuct.IsSellFailure = true
		case `'No stock_transaction found for tx_id = ` + sellerStockID:
			returnStuct.IsSellFailure = true
		default:
			returnStuct.IsSellFailure = true
			returnStuct.IsBuyFailure = true
		}
	}

	log.Println("Time execution: ", time.Since(timeStamp).Seconds(), " seconds")

	// dataStockTransactions, _ := d.StockTransactions().GetAll()
	// for _, v := range *dataStockTransactions {
	// 	jsonData, _ := json.Marshal(v)
	// 	log.Println("Stock Transaction: Data: ", string(jsonData))
	// }

	// dataWalletTransactions, _ := d.WalletTransactions().GetAll()
	// for _, v := range *dataWalletTransactions {
	// 	jsonData, _ := json.Marshal(v)
	// 	log.Println("Wallet Transactions Data: ", string(jsonData))
	// }

	// dataWallet, _ := d.Wallets().GetAll()
	// for _, v := range *dataWallet {
	// 	jsonData, _ := json.Marshal(v)
	// 	log.Println("Wallet Data: ", string(jsonData))
	// }

	// dataUserStocks, _ := d.UserStocks().GetAll()
	// for _, v := range *dataUserStocks {
	// 	jsonData, _ := json.Marshal(v)
	// 	log.Println("User Stocks Data: ", string(jsonData))
	//	}

	return returnStuct, err

}
