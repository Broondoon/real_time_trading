package databaseServiceUserManagement

import (
	databaseService "Shared/database/database-service"
	userStock "Shared/entities/user-stock"
	"Shared/entities/wallet"
	"os"
	"time"

	"gorm.io/gorm"
)

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
}

type DatabaseService struct {
	UserStock databaseService.EntityDataInterface[*userStock.UserStock, *gorm.DB]
	Wallet    databaseService.EntityDataInterface[*wallet.Wallet, *gorm.DB]
	databaseService.DatabaseInterface[*gorm.DB]
}

type NewDatabaseServiceParams struct {
	//UserStockParams *databaseService.NewEntityDataParams // leave nil for default
	//WalletParams    *databaseService.NewEntityDataParams // leave nil for default
	// Only the UserStockParams.NewPostGresDatabaseParams is used. The WalletParams.NewPostGresDatabaseParams is ignored.
}

func NewDatabaseService(params *NewDatabaseServiceParams) DatabaseServiceInterface {
	newDBConnection := databaseService.NewPostGresDatabase(&databaseService.NewPostGresDatabaseParams{})

	cachedUserStock := databaseService.NewCachedEntityData[*userStock.UserStock, *gorm.DB](&databaseService.NewCachedEntityDataParams[*userStock.UserStock, *gorm.DB]{
		RedisAddr:  os.Getenv("REDIS_USER_MANAGEMENT_ADDR"),
		Password:   os.Getenv("REDIS_PASSWORD"),
		DefaultTTL: 5 * time.Minute,
		EntityData: databaseService.NewPostGresEntityData[*userStock.UserStock](
			&databaseService.NewPostGresEntityDataParams{
				Existing: newDBConnection,
			},
		),
	})

	cachedWallet := databaseService.NewCachedEntityData[*wallet.Wallet, *gorm.DB](&databaseService.NewCachedEntityDataParams[*wallet.Wallet, *gorm.DB]{
		RedisAddr:  os.Getenv("REDIS_USER_MANAGEMENT_ADDR"),
		Password:   os.Getenv("REDIS_PASSWORD"),
		DefaultTTL: 5 * time.Minute,
		EntityData: databaseService.NewPostGresEntityData[*wallet.Wallet](
			&databaseService.NewPostGresEntityDataParams{
				Existing: newDBConnection,
			},
		),
	})

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
