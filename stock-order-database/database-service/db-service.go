package databaseServiceStockOrder

import (
	databaseService "Shared/database/database-service"
	"Shared/entities/order"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DatabaseServiceInterface interface {
	databaseService.PostGresDatabaseInterface
	GetStockOrder(orderID string) (order.StockOrderInterface, error)
	GetStockOrders(orderIDs *[]string) (*[]order.StockOrderInterface, error)
	GetInitialStockOrdersForStock(stockID string) (*[]order.StockOrderInterface, error)
	CreateStockOrder(stockOrder order.StockOrderInterface) (order.StockOrderInterface, error)
	UpdateStockOrder(stockOrder order.StockOrderInterface) (order.StockOrderInterface, error)
	DeleteStockOrder(orderID string) (order.StockOrderInterface, error)
}

type DatabaseService struct {
	databaseService.PostGresDatabaseInterface
	StockOrders *[]order.StockOrderInterface
}

type NewDatabaseServiceParams struct {
	*databaseService.NewPostGresDatabaseParams
}

func NewDatabaseService(params NewDatabaseServiceParams) DatabaseServiceInterface {
	db := &DatabaseService{
		PostGresDatabaseInterface: databaseService.NewPostGresDatabase(params.NewPostGresDatabaseParams),
	}
	db.Connect()
	db.GetDatabaseConnection().AutoMigrate(&order.StockOrder{})
	return db
}

func (d *DatabaseService) GetStockOrder(orderID string) (order.StockOrderInterface, error) {
	var order order.StockOrder
	d.GetDatabaseConnection().Where("ID = ?", orderID).First(&order)
	if errors.Is(d.GetDatabaseConnection().Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("stock order not found: %s", orderID)
	}
	order.SetDefaults()
	return &order, nil

}

func (d *DatabaseService) GetStockOrders(orderIDs *[]string) (*[]order.StockOrderInterface, error) {
	var orders []order.StockOrder
	d.GetDatabaseConnection().Where("ID IN ?", orderIDs).Find(&orders)
	if errors.Is(d.GetDatabaseConnection().Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("stock orders not found: %v", orderIDs)
	}
	for _, o := range orders {
		o.SetDefaults()
	}
	converted := make([]order.StockOrderInterface, len(orders))
	for i := range orders {
		converted[i] = &(orders)[i]
	}
	return &converted, nil
}

// Right now, we're just gonna get all stocksOrders for a given stock. Later, we need to limit this to a specific subset of orders.
func (d *DatabaseService) GetInitialStockOrdersForStock(stockID string) (*[]order.StockOrderInterface, error) {
	var orders []order.StockOrder
	d.GetDatabaseConnection().Where("StockID = ? ", stockID).Find(&orders)
	if errors.Is(d.GetDatabaseConnection().Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("stock orders not found")
	}
	for _, o := range orders {
		o.SetDefaults()
	}
	converted := make([]order.StockOrderInterface, len(orders))
	for i := range orders {
		converted[i] = &(orders)[i]
	}
	return &converted, nil
}

// Generated with CHATGPT Because copilot was just screwing around with me. All hail the mighty AI.
func (d *DatabaseService) CreateStockOrder(stockOrder order.StockOrderInterface) (order.StockOrderInterface, error) {
	// Generate a unique order ID. Because a UUID collision is extraordinarily unlikely,
	// most systems skip the existence check altogether.
	// However, if you REALLY want to confirm the ID doesn't already exist
	// (for example, for short random strings, not standard UUIDs), you can do so in a loop.
	var existing order.StockOrder
	for {
		candidateID := generateRandomOrderID()

		// Check if it already exists in the database
		result := d.GetDatabaseConnection().
			Where("id = ?", candidateID).
			First(&existing)

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			stockOrder.SetId(candidateID)
			break
		}

		// If there's another error, handle or return it
		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("error checking existing order: %w", result.Error)
		}
		// If record already found, loop again and generate a new ID
		// (though collisions for standard UUIDs are extremely unlikely)
	}

	// Once we have a unique ID assigned, create the record
	createResult := d.GetDatabaseConnection().Create(stockOrder.(*order.StockOrder))
	if createResult.Error != nil {
		return nil, fmt.Errorf("error creating stock order: %w", createResult.Error)
	}

	return stockOrder, nil
}

func generateRandomOrderID() string {
	// Generate a new UUID as the order ID
	return uuid.New().String()
}

func (d *DatabaseService) UpdateStockOrder(stockOrder order.StockOrderInterface) (order.StockOrderInterface, error) {
	// Update the record in the database
	updateResult := d.GetDatabaseConnection().Save(stockOrder.(*order.StockOrder))
	if updateResult.Error != nil {
		return nil, fmt.Errorf("error updating stock order: %w", updateResult.Error)
	}

	return stockOrder, nil
}

func (d *DatabaseService) DeleteStockOrder(orderID string) (order.StockOrderInterface, error) {
	// Get the order to delete
	stockOrder, err := d.GetStockOrder(orderID)
	if err != nil {
		return nil, fmt.Errorf("error getting stock order: %w", err)
	}

	// Delete the record from the database
	deleteResult := d.GetDatabaseConnection().Delete(stockOrder.(*order.StockOrder), orderID)
	if deleteResult.Error != nil {
		return nil, fmt.Errorf("error deleting stock order: %w", deleteResult.Error)
	}

	return stockOrder, nil
}
