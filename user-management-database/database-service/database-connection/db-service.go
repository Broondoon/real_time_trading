package databaseServiceUserManagement

import (
	databaseService "Shared/database/database-service"
	userStock "Shared/entities/user-stock"
	"Shared/entities/wallet"
)

type UserStockDataServiceInterface interface {
	databaseService.EntityDataInterface[*userStock.UserStock]
}

type WalletDataServiceInterface interface {
	databaseService.EntityDataInterface[*wallet.Wallet]
}

type DatabaseServiceInterface interface {
	databaseService.DatabaseInterface
	UserStocks() UserStockDataServiceInterface
	Wallets() WalletDataServiceInterface
}

type DatabaseService struct {
	UserStock databaseService.EntityDataInterface[*userStock.UserStock]
	Wallet    databaseService.EntityDataInterface[*wallet.Wallet]
	databaseService.DatabaseInterface
}

type NewDatabaseServiceParams struct {
	UserStockParams *databaseService.NewEntityDataParams // leave nil for default
	WalletParams    *databaseService.NewEntityDataParams // leave nil for default
	// Only the UserStockParams.NewPostGresDatabaseParams is used. The WalletParams.NewPostGresDatabaseParams is ignored.
}

func NewDatabaseService(params *NewDatabaseServiceParams) DatabaseServiceInterface {

	if params.UserStockParams == nil {
		params.UserStockParams = &databaseService.NewEntityDataParams{
			NewPostGresDatabaseParams: &databaseService.NewPostGresDatabaseParams{},
		}
	}
	if params.WalletParams == nil {
		params.WalletParams = &databaseService.NewEntityDataParams{
			NewPostGresDatabaseParams: &databaseService.NewPostGresDatabaseParams{},
		}
	}

	var newDBConnection databaseService.PostGresDatabaseInterface
	if params.UserStockParams.Existing != nil {
		newDBConnection = params.UserStockParams.Existing
		if params.WalletParams.Existing == nil {
			params.WalletParams.Existing = newDBConnection
		}
	} else if params.WalletParams.Existing != nil {
		newDBConnection = params.WalletParams.Existing
		params.UserStockParams.Existing = newDBConnection
	} else {
		newDBConnection = databaseService.NewPostGresDatabase(params.UserStockParams.NewPostGresDatabaseParams)
		params.UserStockParams.Existing = newDBConnection
		params.WalletParams.Existing = newDBConnection
	}

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
	db := &DatabaseService{
		UserStock:         databaseService.NewEntityData[*userStock.UserStock](params.UserStockParams),
		Wallet:            databaseService.NewEntityData[*wallet.Wallet](params.WalletParams),
		DatabaseInterface: newDBConnection,
	}

	db.Connect()
	db.UserStocks().GetDatabaseSession().AutoMigrate(&userStock.UserStock{})
	db.Wallets().GetDatabaseSession().AutoMigrate(&wallet.Wallet{})

	return db
}

func (d *DatabaseService) UserStocks() UserStockDataServiceInterface {
	return d.UserStock
}

func (d *DatabaseService) Wallets() WalletDataServiceInterface {
	return d.Wallet
}

func (d *DatabaseService) Connect() {
	d.UserStocks().Connect()
	d.Wallets().Connect()
}

func (d *DatabaseService) Disconnect() {
	d.UserStocks().Disconnect()
	d.Wallets().Disconnect()
}
