package databaseAccessUserManagement

import (
	databaseAccess "Shared/database/database-access"
	userStock "Shared/entities/user-stock"
	"Shared/entities/wallet"
	"Shared/network"
	"errors"
	"fmt"
	"os"
)

type UserStocksDataAccessInterface interface {
	databaseAccess.EntityDataAccessInterface[*userStock.UserStock, userStock.UserStockInterface]
	GetUserStocks(userID string) (*[]userStock.UserStockInterface, error)
	GetUserStocksBulk(userIDs []string, routine func(userID string, userStocks *[]userStock.UserStockInterface, errorCode int)) error
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
}

type DatabaseAccess struct {
	UserStocksDataAccessInterface
	WalletDataAccessInterface
	_networkManager network.NetworkInterface
}

type NewDatabaseAccessParams struct {
	UserStockParams *databaseAccess.NewEntityDataAccessHTTPParams[*userStock.UserStock]
	WalletParams    *databaseAccess.NewEntityDataAccessHTTPParams[*wallet.Wallet]
	Network         network.NetworkInterface
}

func NewDatabaseAccess(params *NewDatabaseAccessParams) DatabaseAccessInterface {
	if params.UserStockParams == nil {
		params.UserStockParams = &databaseAccess.NewEntityDataAccessHTTPParams[*userStock.UserStock]{}
	}

	if params.WalletParams == nil {
		params.WalletParams = &databaseAccess.NewEntityDataAccessHTTPParams[*wallet.Wallet]{}
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

	dba := &DatabaseAccess{
		UserStocksDataAccessInterface: &UserStocksDataAccess{
			EntityDataAccessInterface: databaseAccess.NewEntityDataAccessHTTP[*userStock.UserStock, userStock.UserStockInterface](params.UserStockParams),
		},
		WalletDataAccessInterface: &WalletDataAccess{
			EntityDataAccessInterface: databaseAccess.NewEntityDataAccessHTTP[*wallet.Wallet, wallet.WalletInterface](params.WalletParams),
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

func (d *UserStocksDataAccess) GetUserStocks(userID string) (*[]userStock.UserStockInterface, error) {
	userStocks, err := d.GetByForeignID("user_id", userID)
	if err != nil {
		println("Error fetching user stocks by foreign ID for userID %s: %v\n", userID, err)
		return nil, err
	}
	return userStocks, nil
}

func (d *UserStocksDataAccess) GetUserStocksBulk(userIDs []string, routine func(userID string, userStocks *[]userStock.UserStockInterface, errorCode int)) error {
	userStocks, errList, err := d.GetByForeignIDBulk("user_id", userIDs)
	//lets make a variant which is get by foregin ids. Get back multiple, then perform a function for each userId
	if err != nil {
		println("Error fetching user stocks by foreign ID for userIDs %s: %v\n", userIDs, err)
		return err
	}
	for _, userID := range userIDs {
		userStockslist := []userStock.UserStockInterface{}
		for _, userStock := range *userStocks {
			if userStock.GetUserID() == userID {
				userStockslist = append(userStockslist, userStock)
			}
		}
		go routine(userID, &userStockslist, errList[userID])
	}
	return nil
}

func (d *WalletDataAccess) AddMoneyToWallet(userID string, amount float64) error {
	fmt.Printf("DEBUG: AddMoneyToWallet called for userID=%s with amount=%f\n", userID, amount)

	walletList, err := d.GetByForeignID("user_id", userID)
	if err != nil {
		fmt.Printf("DEBUG: Error retrieving wallet for userID=%s: %v\n", userID, err)
		return err
	}
	fmt.Printf("DEBUG: Retrieved %d wallet(s) for userID=%s\n", len(*walletList), userID)

	if len(*walletList) == 0 {
		fmt.Printf("DEBUG: No wallet found for userID=%s\n", userID)
		return errors.New("no wallet found for user")
	}

	wallet := (*walletList)[0]
	oldBalance := wallet.GetBalance()
	newBalance := oldBalance + amount
	fmt.Printf("DEBUG: Updating wallet for userID=%s: old balance=%f, new balance=%f\n", userID, oldBalance, newBalance)

	wallet.UpdateBalance(amount)
	err = d.Update(wallet)
	if err != nil {
		fmt.Printf("DEBUG: Error updating wallet for userID=%s: %v\n", userID, err)
		return err
	}

	fmt.Printf("DEBUG: Successfully updated wallet for userID=%s\n", userID)
	return nil
}

func (d *WalletDataAccess) GetWalletBalance(userID string) (float64, error) {
	walletList, err := d.GetByForeignID("user_id", userID)
	if err != nil {
		fmt.Printf("[DEBUG] Error fetching wallet by foreign ID for userID %s: %v\n", userID, err)
		return 0, err
	}
	fmt.Printf("[DEBUG] Retrieved walletList for userID %s: %v\n", userID, walletList)
	if len(*walletList) == 0 {
		fmt.Printf("[DEBUG] No wallet found for userID: %s\n", userID)
		return 0, errors.New("no wallet found for user")
	}
	wallet := (*walletList)[0]
	return wallet.GetBalance(), nil
}
