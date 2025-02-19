package databaseAccessTransaction

import (
	databaseAccess "Shared/database/database-access"
	"Shared/entities/transaction"
	"Shared/network"
	"os"
)

type StockTransactionDataAccessInterface = databaseAccess.EntityDataAccessInterface[*transaction.StockTransaction, transaction.StockTransactionInterface]
type WalletTransactionDataAccessInterface = databaseAccess.EntityDataAccessInterface[*transaction.WalletTransaction, transaction.WalletTransactionInterface]

type DatabaseAccessInterface interface {
	databaseAccess.DatabaseAccessInterface
	StockTransaction() StockTransactionDataAccessInterface
	WalletTransaction() WalletTransactionDataAccessInterface
}

type DatabaseAccess struct {
	StockTransactionDataAccessInterface
	WalletTransactionDataAccessInterface
	_networkManager network.NetworkInterface
}

type NewDatabaseAccessParams struct {
	StockTransactionParams  *databaseAccess.NewEntityDataAccessHTTPParams[*transaction.StockTransaction]
	WalletTransactionParams *databaseAccess.NewEntityDataAccessHTTPParams[*transaction.WalletTransaction]
	Network                 network.NetworkInterface
}

func NewDatabaseAccess(params *NewDatabaseAccessParams) DatabaseAccessInterface {
	if params.StockTransactionParams == nil {
		params.StockTransactionParams = &databaseAccess.NewEntityDataAccessHTTPParams[*transaction.StockTransaction]{}
	}

	if params.WalletTransactionParams == nil {
		params.WalletTransactionParams = &databaseAccess.NewEntityDataAccessHTTPParams[*transaction.WalletTransaction]{}
	}

	if params.Network == nil {
		panic("No network provided")
	}

	if params.StockTransactionParams.Client == nil {
		params.StockTransactionParams.Client = params.Network.Transactions()
	}
	if params.StockTransactionParams.DefaultRoute == "" {
		params.StockTransactionParams.DefaultRoute = os.Getenv("TRANSACTION_DATABASE_SERVICE_STOCK_ROUTE")
	}
	if params.WalletTransactionParams.Client == nil {
		params.WalletTransactionParams.Client = params.Network.Transactions()
	}
	if params.WalletTransactionParams.DefaultRoute == "" {
		params.WalletTransactionParams.DefaultRoute = os.Getenv("TRANSACTION_DATABASE_SERVICE_WALLET_ROUTE")
	}

	dba := &DatabaseAccess{
		StockTransactionDataAccessInterface:  databaseAccess.NewEntityDataAccessHTTP[*transaction.StockTransaction, transaction.StockTransactionInterface](params.StockTransactionParams),
		WalletTransactionDataAccessInterface: databaseAccess.NewEntityDataAccessHTTP[*transaction.WalletTransaction, transaction.WalletTransactionInterface](params.WalletTransactionParams),
		_networkManager:                      params.Network,
	}

	dba.Connect()
	return dba
}

func (d *DatabaseAccess) Connect() {
}

func (d *DatabaseAccess) Disconnect() {
}

func (d *DatabaseAccess) StockTransaction() StockTransactionDataAccessInterface {
	return d.StockTransactionDataAccessInterface
}

func (d *DatabaseAccess) WalletTransaction() WalletTransactionDataAccessInterface {
	return d.WalletTransactionDataAccessInterface
}
