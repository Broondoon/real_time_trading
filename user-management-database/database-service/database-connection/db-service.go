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
	UserStock UserStockDataServiceInterface
	Wallet    WalletDataServiceInterface
	databaseService.DatabaseInterface
}

type NewDatabaseServiceParams struct {
	UserStockParams *databaseService.NewEntityDataParams // leave nil for default
	WalletParams    *databaseService.NewEntityDataParams // leave nil for default
	//Only the UserStockParams.NewPostGresDatabaseParams is used. The WalletParams.NewPostGresDatabaseParams is ignored.
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
	d.UserStocks().Connect()
}

func (d *DatabaseService) Disconnect() {
	d.UserStocks().Disconnect()
	d.UserStocks().Disconnect()
}

// type DatabaseServiceInterface interface {
// 	CreateWallet(wallet *wallet.Wallet) error
// 	GetWalletByUserID(userID string) (*wallet.Wallet, error)
// 	AddMoneyToWallet(userID string, amount float64) error
// 	GetWalletBalance(userID string) (float64, error)

// 	CreateUserStock(userStock *userStock.UserStock) error
// 	GetUserStocksByUserID(userID string) ([]userStock.UserStock, error)
// }

// type DatabaseService struct {
// 	DB *gorm.DB
// }

// func NewDatabaseService(db *gorm.DB) DatabaseServiceInterface {
// 	return &DatabaseService{DB: db}
// }

// func (d *DatabaseService) CreateWallet(wallet *wallet.Wallet) error {
// 	return d.DB.Create(wallet).Error
// }

// func (d *DatabaseService) GetWalletByUserID(userID string) (*wallet.Wallet, error) {
// 	var wallet wallet.Wallet
// 	err := d.DB.First(&wallet, "user_id = ?", userID).Error
// 	return &wallet, err
// }

// func (d *DatabaseService) AddMoneyToWallet(userID string, amount float64) error {
// 	wallet, err := d.GetWalletByUserID(userID)
// 	if err != nil {
// 		return err
// 	}
// 	wallet.Balance += amount
// 	return d.DB.Save(wallet).Error
// }

// func (d *DatabaseService) GetWalletBalance(userID string) (float64, error) {
// 	wallet, err := d.GetWalletByUserID(userID)
// 	if err != nil {
// 		return 0, err
// 	}
// 	return wallet.Balance, nil
// }

// func (d *DatabaseService) CreateUserStock(userStock *userStock.UserStock) error {
// 	return d.DB.Create(userStock).Error
// }

// func (d *DatabaseService) GetUserStocksByUserID(userID string) ([]userStock.UserStock, error) {
// 	var stocks []userStock.UserStock
// 	err := d.DB.Find(&stocks, "user_id = ?", userID).Error
// 	return stocks, err
// }
