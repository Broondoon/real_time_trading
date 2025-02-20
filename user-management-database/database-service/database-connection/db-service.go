package databaseServiceUserManagement

import (
	userStock "Shared/entities/user-stock"
	"Shared/entities/wallet"

	"gorm.io/gorm"
)

// DatabaseServiceInterface combines Wallet and UserStock operations
type DatabaseServiceInterface interface {
	// Wallet operations
	CreateWallet(wallet *wallet.Wallet) error
	GetWalletByUserID(userID string) (*wallet.Wallet, error)
	AddMoneyToWallet(userID string, amount float64) error
	GetWalletBalance(userID string) (float64, error)

	// UserStock operations
	CreateUserStock(userStock *userStock.UserStock) error
	GetUserStocksByUserID(userID string) ([]userStock.UserStock, error)
}

// DatabaseService implements DatabaseServiceInterface
type DatabaseService struct {
	DB *gorm.DB
}

// NewDatabaseService initializes the service with a GORM database instance
func NewDatabaseService(db *gorm.DB) DatabaseServiceInterface {
	return &DatabaseService{DB: db}
}

// -------------------- Wallet Methods --------------------

// CreateWallet creates a new wallet
func (d *DatabaseService) CreateWallet(wallet *wallet.Wallet) error {
	return d.DB.Create(wallet).Error
}

// GetWalletByUserID retrieves a wallet by user ID
func (d *DatabaseService) GetWalletByUserID(userID string) (*wallet.Wallet, error) {
	var wallet wallet.Wallet
	err := d.DB.First(&wallet, "user_id = ?", userID).Error
	return &wallet, err
}

// AddMoneyToWallet adds money to a user's wallet
func (d *DatabaseService) AddMoneyToWallet(userID string, amount float64) error {
	wallet, err := d.GetWalletByUserID(userID)
	if err != nil {
		return err
	}
	wallet.Balance += amount
	return d.DB.Save(wallet).Error
}

// GetWalletBalance retrieves the wallet balance for a user
func (d *DatabaseService) GetWalletBalance(userID string) (float64, error) {
	wallet, err := d.GetWalletByUserID(userID)
	if err != nil {
		return 0, err
	}
	return wallet.Balance, nil
}

// -------------------- UserStock Methods --------------------

// CreateUserStock creates a new user stock record
func (d *DatabaseService) CreateUserStock(userStock *userStock.UserStock) error {
	return d.DB.Create(userStock).Error
}

// GetUserStocksByUserID retrieves user stocks for a specific user
func (d *DatabaseService) GetUserStocksByUserID(userID string) ([]userStock.UserStock, error) {
	var stocks []userStock.UserStock
	err := d.DB.Find(&stocks, "user_id = ?", userID).Error
	return stocks, err
}
