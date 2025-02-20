package databaseAccessUserManagement

import (
	databaseAccess "Shared/database/database-access"
	userStock "Shared/entities/user-stock"
	databaseServiceUserManagement "databaseServiceUserManagement/database-connection"
)

type UserStockDataAccessInterface interface {
	databaseAccess.DatabaseAccessInterface
	GetUserStocks(userID string) ([]userStock.UserStockInterface, error)
}

type UserStockDatabaseAccess struct {
	databaseAccess.EntityDataAccessInterface[*userStock.UserStock, userStock.UserStockInterface]
	TEMPCONNECTION databaseServiceUserManagement.DatabaseServiceInterface
}

func NewUserStockDatabaseAccess(service databaseServiceUserManagement.DatabaseServiceInterface) UserStockDataAccessInterface {
	return &UserStockDatabaseAccess{
		EntityDataAccessInterface: databaseAccess.NewEntityDataAccess[*userStock.UserStock, userStock.UserStockInterface](
			&databaseAccess.NewEntityDataAccessParams[*userStock.UserStock]{}),
		TEMPCONNECTION: service,
	}
}

func (d *UserStockDatabaseAccess) GetUserStocks(userID string) ([]userStock.UserStockInterface, error) {
	stocks, err := d.TEMPCONNECTION.GetUserStocksByUserID(userID)
	if err != nil {
		return nil, err
	}

	convertedStocks := make([]userStock.UserStockInterface, len(stocks))
	for i, stock := range stocks {
		convertedStocks[i] = &stock
	}
	return convertedStocks, nil
}
