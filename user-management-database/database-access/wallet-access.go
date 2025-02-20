package databaseAccessUserManagement

import (
	databaseAccess "Shared/database/database-access"
	"Shared/entities/wallet"
	databaseServiceUserManagement "databaseServiceUserManagement/database-connection"
)

type WalletDataAccessInterface interface {
	databaseAccess.DatabaseAccessInterface
	AddMoneyToWallet(userID string, amount float64) error
	GetWalletBalance(userID string) (float64, error)
}

type WalletDatabaseAccess struct {
	databaseAccess.EntityDataAccessInterface[*wallet.Wallet, wallet.WalletInterface]
	TEMPCONNECTION databaseServiceUserManagement.DatabaseServiceInterface
}

func NewWalletDatabaseAccess(service databaseServiceUserManagement.DatabaseServiceInterface) WalletDataAccessInterface {
	return &WalletDatabaseAccess{
		EntityDataAccessInterface: databaseAccess.NewEntityDataAccess[*wallet.Wallet, wallet.WalletInterface](
			&databaseAccess.NewEntityDataAccessParams[*wallet.Wallet]{}),
		TEMPCONNECTION: service,
	}
}

func (d *WalletDatabaseAccess) AddMoneyToWallet(userID string, amount float64) error {
	return d.TEMPCONNECTION.AddMoneyToWallet(userID, amount)
}

func (d *WalletDatabaseAccess) GetWalletBalance(userID string) (float64, error) {
	return d.TEMPCONNECTION.GetWalletBalance(userID)
}
