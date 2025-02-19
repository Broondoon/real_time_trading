package main

import (
	"Shared/entities/entity"
	"Shared/entities/order"
	databaseAccessStockOrder "databaseAccessStockOrder"
	"fmt"
	"time"
	//"Shared/network"
)

func main() {
	//Create a new Stock Order
	so := order.New(order.NewStockOrderParams{
		NewEntityParams: entity.NewEntityParams{
			ID:           "so1",
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		StockID:   "e",
		Quantity:  0,
		Price:     0.0,
		OrderType: "MARKET",
		IsBuy:     true,
	})
	// Print the Stock Order
	fmt.Print("Stock Order: ")
	fmt.Println(so.GetId())
	fmt.Println(so.GetDateCreated())
	fmt.Println(so.GetDateModified())
	fmt.Println(so.GetStockID())
	fmt.Println(so.GetQuantity())
	fmt.Println(so.GetPrice())
	fmt.Println(so.GetOrderType())
	fmt.Println(so.GetIsBuy())

	//testing database.
	_databaseManager := databaseAccessStockOrder.NewDatabaseAccess(&databaseAccessStockOrder.NewDatabaseAccessParams{})

	fmt.Println("Database Test")
	fmt.Println("Testing create Stock Order: ")
	_databaseManager.Create(so)
	fmt.Println("Stock Order Created with ID: ", so.GetId())
	so2 := _databaseManager.GetByID(so.GetId())
	fmt.Print("Testing get Stock Order: ")
	fmt.Println(so2.GetId())
	fmt.Println(so2.GetDateCreated())
	fmt.Println(so2.GetDateModified())
	fmt.Println(so2.GetStockID())
	fmt.Println(so2.GetQuantity())
	fmt.Println(so2.GetPrice())
	fmt.Println(so2.GetOrderType())
	fmt.Println(so2.GetIsBuy())

	fmt.Print("Testing group get Stock Orders: ")
	idList := []string{"so1", so.GetId()}
	so5 := _databaseManager.GetByIDs(idList)
	for _, so6 := range *so5 {
		fmt.Println(so6.GetId())
		fmt.Println(so6.GetDateCreated())
		fmt.Println(so6.GetDateModified())
		fmt.Println(so6.GetStockID())
		fmt.Println(so6.GetQuantity())
		fmt.Println(so6.GetPrice())
		fmt.Println(so6.GetOrderType())
		fmt.Println(so6.GetIsBuy())
	}

	fmt.Println("Testing update Stock Order: ")
	so.SetIsBuy(false)
	_databaseManager.Update(so)
	so4 := _databaseManager.GetByID(so.GetId())
	fmt.Print("Stock Order: ")
	fmt.Println(so4.GetId())
	fmt.Println(so4.GetDateCreated())
	fmt.Println(so4.GetDateModified())
	fmt.Println(so4.GetStockID())
	fmt.Println(so4.GetQuantity())
	fmt.Println(so4.GetPrice())
	fmt.Println(so4.GetOrderType())
	fmt.Println(so4.GetIsBuy())

	fmt.Println("Testing delete Stock Order: ")
	_databaseManager.Delete(so.GetId())
	_databaseManager.GetByID(so.GetId())

}
