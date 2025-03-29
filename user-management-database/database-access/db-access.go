package databaseAccessUserManagement

import (
	databaseAccess "Shared/database/database-access"
	"Shared/entities/transaction"
	userStock "Shared/entities/user-stock"
	"Shared/entities/wallet"
	"Shared/network"
	"encoding/json"
	"errors"
	"log"
	"os"
)

type WalletTransactionDataAccessInterface = databaseAccess.EntityDataAccessInterface[*transaction.WalletTransaction, transaction.WalletTransactionInterface]

type WalletTransactionDataAccess struct {
	WalletTransactionDataAccessInterface
}

type StockTransactionDataAccessInterface = databaseAccess.EntityDataAccessInterface[*transaction.StockTransaction, transaction.StockTransactionInterface]

type StockTransactionDataAccess struct {
	StockTransactionDataAccessInterface
}

type UserStocksDataAccessInterface interface {
	databaseAccess.EntityDataAccessInterface[*userStock.UserStock, userStock.UserStockInterface]
	GetUserStocks(userID string) (*[]userStock.UserStockInterface, error)
	GetUserStocksBulk(userIDs []string, stockId string, routine func(userID string, userStocks *[]userStock.UserStockInterface, errorCode int)) error
}

type UserStocksDataAccess struct {
	databaseAccess.EntityDataAccessInterface[*userStock.UserStock, userStock.UserStockInterface]
}

type WalletDataAccessInterface interface {
	databaseAccess.EntityDataAccessInterface[*wallet.Wallet, wallet.WalletInterface]
	AddMoneyToWallet(userID string, amount float64) error
	GetWalletBalance(userID string) (float64, error)
}

type WalletDataAccess struct {
	databaseAccess.EntityDataAccessInterface[*wallet.Wallet, wallet.WalletInterface]
}

type DatabaseAccessInterface interface {
	databaseAccess.DatabaseAccessInterface
	UserStock() UserStocksDataAccessInterface
	Wallet() WalletDataAccessInterface
	StockTransaction() StockTransactionDataAccessInterface
	WalletTransaction() WalletTransactionDataAccessInterface
	ExecuteOrder(buyerStockID string, sellerStockID string, buyerID string, sellerID string, stockID string, stockName string, stockPrice float64, quantity int) (network.ExecutorToMatchingEngineJSON, error)
}

type DatabaseAccess struct {
	UserStocksDataAccessInterface
	WalletDataAccessInterface
	StockTransactionDataAccessInterface
	WalletTransactionDataAccessInterface
	_networkManager network.NetworkInterface
}

type NewDatabaseAccessParams struct {
	UserStockParams         *databaseAccess.NewEntityDataAccessNetworkParams[*userStock.UserStock]
	WalletParams            *databaseAccess.NewEntityDataAccessNetworkParams[*wallet.Wallet]
	StockTransactionParams  *databaseAccess.NewEntityDataAccessNetworkParams[*transaction.StockTransaction]
	WalletTransactionParams *databaseAccess.NewEntityDataAccessNetworkParams[*transaction.WalletTransaction]
	Network                 network.NetworkInterface
}

func NewDatabaseAccess(params *NewDatabaseAccessParams) DatabaseAccessInterface {
	if params.UserStockParams == nil {
		params.UserStockParams = &databaseAccess.NewEntityDataAccessNetworkParams[*userStock.UserStock]{}
	}

	if params.WalletParams == nil {
		params.WalletParams = &databaseAccess.NewEntityDataAccessNetworkParams[*wallet.Wallet]{}
	}

	if params.StockTransactionParams == nil {
		params.StockTransactionParams = &databaseAccess.NewEntityDataAccessNetworkParams[*transaction.StockTransaction]{}
	}

	if params.WalletTransactionParams == nil {
		params.WalletTransactionParams = &databaseAccess.NewEntityDataAccessNetworkParams[*transaction.WalletTransaction]{}
	}

	if params.Network == nil {
		panic("No network provided")
	}

	if params.UserStockParams.Client == nil {
		params.UserStockParams.Client = params.Network.UserManagementDatabase()
	}
	if params.UserStockParams.DefaultRoute == "" {
		params.UserStockParams.DefaultRoute = os.Getenv("USER_MANAGEMENT_SERVICE_USER_STOCK_ROUTE")
	}
	if params.WalletParams.Client == nil {
		params.WalletParams.Client = params.Network.UserManagementDatabase()
	}
	if params.WalletParams.DefaultRoute == "" {
		params.WalletParams.DefaultRoute = os.Getenv("USER_MANAGEMENT_SERVICE_WALLET_ROUTE")
	}

	if params.UserStockParams.Parser == nil {
		params.UserStockParams.Parser = userStock.Parse
	}
	if params.WalletParams.Parser == nil {
		params.WalletParams.Parser = wallet.Parse
	}
	if params.UserStockParams.ParserList == nil {
		params.UserStockParams.ParserList = userStock.ParseList
	}
	if params.WalletParams.ParserList == nil {
		params.WalletParams.ParserList = wallet.ParseList
	}

	if params.StockTransactionParams.Client == nil {
		params.StockTransactionParams.Client = params.Network.UserManagementDatabase()
	}
	if params.StockTransactionParams.DefaultRoute == "" {
		params.StockTransactionParams.DefaultRoute = os.Getenv("TRANSACTION_DATABASE_SERVICE_STOCK_ROUTE")
	}
	if params.WalletTransactionParams.Client == nil {
		params.WalletTransactionParams.Client = params.Network.UserManagementDatabase()
	}
	if params.WalletTransactionParams.DefaultRoute == "" {
		params.WalletTransactionParams.DefaultRoute = os.Getenv("TRANSACTION_DATABASE_SERVICE_WALLET_ROUTE")
	}

	if params.StockTransactionParams.Parser == nil {
		params.StockTransactionParams.Parser = transaction.ParseStockTransaction
	}
	if params.WalletTransactionParams.Parser == nil {
		params.WalletTransactionParams.Parser = transaction.ParseWalletTransaction
	}
	if params.StockTransactionParams.ParserList == nil {
		params.StockTransactionParams.ParserList = transaction.ParseStockTransactionList
	}
	if params.WalletTransactionParams.ParserList == nil {
		params.WalletTransactionParams.ParserList = transaction.ParseWalletTransactionList
	}

	dba := &DatabaseAccess{
		UserStocksDataAccessInterface: &UserStocksDataAccess{
			EntityDataAccessInterface: databaseAccess.NewEntityDataAccessNetwork[*userStock.UserStock, userStock.UserStockInterface](params.UserStockParams),
		},
		WalletDataAccessInterface: &WalletDataAccess{
			EntityDataAccessInterface: databaseAccess.NewEntityDataAccessNetwork[*wallet.Wallet, wallet.WalletInterface](params.WalletParams),
		},
		StockTransactionDataAccessInterface: &StockTransactionDataAccess{
			StockTransactionDataAccessInterface: databaseAccess.NewEntityDataAccessNetwork[*transaction.StockTransaction, transaction.StockTransactionInterface](params.StockTransactionParams),
		},
		WalletTransactionDataAccessInterface: &WalletTransactionDataAccess{
			WalletTransactionDataAccessInterface: databaseAccess.NewEntityDataAccessNetwork[*transaction.WalletTransaction, transaction.WalletTransactionInterface](params.WalletTransactionParams),
		},
		_networkManager: params.Network,
	}

	dba.Connect()
	return dba
}

func (d *DatabaseAccess) Connect() {
}

func (d *DatabaseAccess) Disconnect() {
}

func (d *DatabaseAccess) UserStock() UserStocksDataAccessInterface {
	return d.UserStocksDataAccessInterface
}

func (d *DatabaseAccess) Wallet() WalletDataAccessInterface {
	return d.WalletDataAccessInterface
}

func (d *DatabaseAccess) StockTransaction() StockTransactionDataAccessInterface {
	return d.StockTransactionDataAccessInterface
}

func (d *DatabaseAccess) WalletTransaction() WalletTransactionDataAccessInterface {
	return d.WalletTransactionDataAccessInterface
}

func (d *UserStocksDataAccess) GetUserStocks(userID string) (*[]userStock.UserStockInterface, error) {
	userStocks, err := d.GetByForeignID("UserID", userID)
	if err != nil {
		log.Println("Error fetching user stocks by foreign ID for userID %s: %v\n", userID, err)
		return nil, err
	}
	return userStocks, nil
}

func (d *UserStocksDataAccess) GetUserStocksBulk(userIDs []string, stockId string, routine func(userID string, userStocks *[]userStock.UserStockInterface, errorCode int)) error {

	if len(userIDs) == 0 {
		log.Printf("DEBUG: GetUserStocksBulk called with empty userIDs\n")
		return nil
	}
	//log.Println("DEBUG: GetUserStocksBulk called for userIDs %s\n", userIDs)
	var (
		userStocks *[]userStock.UserStockInterface
		errList    = make(map[string]int)
		err        error
	)
	if stockId != "" {
		userStocks, errList, err = d.GetByFilteredForeignIDBulk("UserID", userIDs, "StockID", stockId)
	} else {
		userStocks, errList, err = d.GetByForeignIDBulk("UserID", userIDs)
	}
	//lets make a variant which is get by foregin ids. Get back multiple, then perform a function for each userId
	if err != nil {
		log.Printf("Error fetching user stocks by foreign ID for userIDs %s: %v\n", userIDs, err)
		return err
	}
	for _, userID := range userIDs {
		userStockslist := []userStock.UserStockInterface{}
		for _, userStock := range *userStocks {
			if userStock.GetUserIDString() == userID {
				userStockslist = append(userStockslist, userStock)
			}
		}
		go routine(userID, &userStockslist, errList[userID])
	}
	return nil
}

func (d *WalletDataAccess) AddMoneyToWallet(userID string, amount float64) error {
	//log.Printf("DEBUG: AddMoneyToWallet called for userID=%s with amount=%f\n", userID, amount)

	walletList, err := d.GetByForeignID("UserID", userID)
	if err != nil {
		log.Printf("DEBUG: Error retrieving wallet for userID=%s: %v\n", userID, err)
		return err
	}
	//log.Printf("DEBUG: Retrieved %d wallet(s) for userID=%s\n", len(*walletList), userID)

	if len(*walletList) == 0 {
		log.Printf("DEBUG: No wallet found for userID=%s\n", userID)
		return errors.New("no wallet found for user")
	}

	wallet := (*walletList)[0]
	//oldBalance := wallet.GetBalance()
	//newBalance := oldBalance + amount
	//log.Printf("DEBUG: Updating wallet for userID=%s: old balance=%f, new balance=%f\n", userID, oldBalance, newBalance)

	wallet.UpdateBalance(amount)
	err = d.Update(wallet)
	if err != nil {
		log.Printf("DEBUG: Error updating wallet for userID=%s: %v\n", userID, err)
		return err
	}

	//log.Printf("DEBUG: Successfully updated wallet for userID=%s\n", userID)
	return nil
}

func (d *WalletDataAccess) GetWalletBalance(userID string) (float64, error) {
	walletList, err := d.GetByForeignID("UserID", userID)
	if err != nil {
		log.Printf("[DEBUG] Error fetching wallet by foreign ID for userID %s: %v\n", userID, err)
		return 0, err
	}
	//log.Printf("[DEBUG] Retrieved walletList for userID %s: %v\n", userID, walletList)
	if len(*walletList) == 0 {
		log.Printf("[DEBUG] No wallet found for userID: %s\n", userID)
		return 0, errors.New("no wallet found for user")
	}
	wallet := (*walletList)[0]
	return wallet.GetBalance(), nil
}

func (d *DatabaseAccess) ExecuteOrder(buyerStockID string, sellerStockID string, buyerID string, sellerID string, stockID string, stockName string, stockPrice float64, quantity int) (network.ExecutorToMatchingEngineJSON, error) {
	log.Println("Access: Executing order")
	returnJson, err := d._networkManager.UserManagementDatabase().Post("execute-order", network.MatchingEngineToExecutionJSON{
		BuyerID:     buyerID,
		SellerID:    sellerID,
		StockID:     stockID,
		BuyOrderID:  buyerStockID,
		SellOrderID: sellerStockID,
		StockPrice:  stockPrice,
		Quantity:    quantity,
		Name:        stockName,
	})
	var returnData network.ExecutorToMatchingEngineJSON
	if err != nil {
		log.Printf("Error executing order: %v\n", err)
		return returnData, err
	}
	err = json.Unmarshal(returnJson, &returnData)
	log.Println("Returned executed order")
	return returnData, err
}
