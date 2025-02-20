package databaseServiceUserManagement

import (
	userStock "Shared/entities/user-stock"
	"Shared/entities/wallet"

	"gorm.io/gorm"
)

type DatabaseServiceInterface interface {
	CreateWallet(wallet *wallet.Wallet) error
	GetWalletByUserID(userID string) (*wallet.Wallet, error)
	AddMoneyToWallet(userID string, amount float64) error
	GetWalletBalance(userID string) (float64, error)

	CreateUserStock(userStock *userStock.UserStock) error
	GetUserStocksByUserID(userID string) ([]userStock.UserStock, error)
}

type DatabaseService struct {
	DB *gorm.DB
}

func NewDatabaseService(db *gorm.DB) DatabaseServiceInterface {
	return &DatabaseService{DB: db}
}

func (d *DatabaseService) CreateWallet(wallet *wallet.Wallet) error {
	return d.DB.Create(wallet).Error
}

func (d *DatabaseService) GetWalletByUserID(userID string) (*wallet.Wallet, error) {
	var wallet wallet.Wallet
	err := d.DB.First(&wallet, "user_id = ?", userID).Error
	return &wallet, err
}

func (d *DatabaseService) AddMoneyToWallet(userID string, amount float64) error {
	wallet, err := d.GetWalletByUserID(userID)
	if err != nil {
		return err
	}
	wallet.Balance += amount
	return d.DB.Save(wallet).Error
}

func (d *DatabaseService) GetWalletBalance(userID string) (float64, error) {
	wallet, err := d.GetWalletByUserID(userID)
	if err != nil {
		return 0, err
	}
	return wallet.Balance, nil
}

func (d *DatabaseService) CreateUserStock(userStock *userStock.UserStock) error {
	return d.DB.Create(userStock).Error
}

func (d *DatabaseService) GetUserStocksByUserID(userID string) ([]userStock.UserStock, error) {
	var stocks []userStock.UserStock
	err := d.DB.Find(&stocks, "user_id = ?", userID).Error
	return stocks, err
}
