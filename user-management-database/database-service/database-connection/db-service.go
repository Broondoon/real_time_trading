package databaseServiceUserManagement

import (
	"os"
	"time"

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

	underlyingUserStock := databaseService.NewEntityData[*userStock.UserStock](params.UserStockParams)
	underlyingWallet := databaseService.NewEntityData[*wallet.Wallet](params.WalletParams)

	// Wrap the underlying services with the caching layer.
	cachedUserStock := databaseService.NewCachedEntityData[*userStock.UserStock](
		underlyingUserStock,
		os.Getenv("REDIS_ADDR"),     // e.g., "redis:6379" from your Docker Compose network
		os.Getenv("REDIS_PASSWORD"), // leave empty if no password
		5*time.Minute,               // default TTL for cache entries
	)
	cachedWallet := databaseService.NewCachedEntityData[*wallet.Wallet](
		underlyingWallet,
		os.Getenv("REDIS_ADDR"),
		os.Getenv("REDIS_PASSWORD"),
		5*time.Minute,
	)

	db := &DatabaseService{
		UserStock:         cachedUserStock,
		Wallet:            cachedWallet,
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
