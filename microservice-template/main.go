package main

import (
	"Shared/entities/entity"
	"Shared/entities/order"
	"Shared/entities/stock"
	userStock "Shared/entities/user-stock"
	"Shared/entities/wallet"
	"Shared/network"
	"databaseAccessUserManagement"
	"time"

	"github.com/google/uuid"
	//"Shared/network"
)

func main() {
	var err error
	var val []byte

	networkManager := network.NewNetwork()

	// val, err = networkManager.Transactions().Get("transaction/getStockTransactions", nil)
	// if err != nil {
	// 	panic(err)
	// }
	// println(string(val))
	// println("Stock Transactions gotten")

	ea := databaseAccessUserManagement.NewDatabaseAccess(&databaseAccessUserManagement.NewDatabaseAccessParams{
		Network: networkManager,
	})

	test, err := ea.Wallet().Create(wallet.New(wallet.NewWalletParams{
		NewEntityParams: entity.NewEntityParams{
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		UserID:  "6fd2fc6b-9142-4777-8b30-575ff6fa2460",
		Balance: 100000,
	}))
	if err != nil {
		println("Error creating wallet: ", err)
		panic(err)
	}
	walletOutput, err := test.ToJSON()
	println("Wallet created: ", string(walletOutput))

	testArray, err := ea.Wallet().GetByForeignID("user_id", "6fd2fc6b-9142-4777-8b30-575ff6fa2460")
	if err != nil {
		println("Error getting wallet: ", err)
		panic(err)
	}
	println(testArray)
	println("Wallet gotten")

	//Create a new Stock
	newStock1 := stock.New(stock.NewStockParams{
		NewEntityParams: entity.NewEntityParams{
			ID:           uuid.New().String(),
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		Name: "Apple",
	})
	newStock2 := stock.New(stock.NewStockParams{
		NewEntityParams: entity.NewEntityParams{
			ID:           uuid.New().String(),
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		Name: "google",
	})

	val, err = networkManager.Stocks().Post("setup/createStock", newStock1)
	if err != nil {
		panic(err)
	}

	val, err = networkManager.Stocks().Post("setup/createStock", newStock2)
	if err != nil {
		panic(err)
	}
	println("Stock Created")

	ea.UserStock().Create(userStock.New(userStock.NewUserStockParams{
		NewEntityParams: entity.NewEntityParams{
			ID: uuid.New().String(),
		},
		UserID:   "6fd2fc6b-9142-4777-8b30-575ff6fa2460",
		StockID:  newStock1.GetId(),
		Quantity: 100,
	}))

	ea.UserStock().Create(userStock.New(userStock.NewUserStockParams{
		NewEntityParams: entity.NewEntityParams{
			ID: uuid.New().String(),
		},
		UserID:   "6fd2fc6b-9142-4777-8b30-575ff6fa2460",
		StockID:  newStock2.GetId(),
		Quantity: 100,
	}))

	d, err := ea.UserStock().GetByForeignID("user_id", "6fd2fc6b-9142-4777-8b30-575ff6fa2460")
	if err != nil {
		println("Error getting user stock: ", err)
		panic(err)
	}
	for _, us := range *d {
		da, err := us.ToJSON()
		if err != nil {
			println("Error converting user stock to json: ", err)
			panic(err)
		}
		println("User Stock gotten: ", string(da))
	}

	//Create a new Stock Order
	so1 := order.New(order.NewStockOrderParams{
		NewEntityParams: entity.NewEntityParams{
			ID:           "so1",
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		StockID:   newStock1.GetId(),
		Quantity:  5,
		Price:     7.5,
		OrderType: "LIMIT",
		IsBuy:     false,
	})

	// so2 := order.New(order.NewStockOrderParams{
	// 	NewEntityParams: entity.NewEntityParams{
	// 		ID:           "so2",
	// 		DateCreated:  time.Now(),
	// 		DateModified: time.Now(),
	// 	},
	// 	StockID:   newStock1.GetId(),
	// 	Quantity:  5,
	// 	Price:     7.5,
	// 	OrderType: "LIMIT",
	// 	IsBuy:     false,
	// })

	// so3 := order.New(order.NewStockOrderParams{

	// 	NewEntityParams: entity.NewEntityParams{
	// 		ID:           "so3",
	// 		DateCreated:  time.Now(),
	// 		DateModified: time.Now(),
	// 	},
	// 	StockID:   newStock1.GetId(),
	// 	Quantity:  5,
	// 	Price:     8.5,
	// 	OrderType: "LIMIT",
	// 	IsBuy:     false,
	// })

	// so4 := order.New(order.NewStockOrderParams{
	// 	NewEntityParams: entity.NewEntityParams{
	// 		ID:           "so4",
	// 		DateCreated:  time.Now(),
	// 		DateModified: time.Now(),
	// 	},
	// 	StockID:   newStock1.GetId(),
	// 	Quantity:  5,
	// 	Price:     6.5,
	// 	OrderType: "LIMIT",
	// 	IsBuy:     false,
	// })

	so5 := order.New(order.NewStockOrderParams{
		NewEntityParams: entity.NewEntityParams{
			ID:           "so5",
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		StockID:   newStock1.GetId(),
		Quantity:  5,
		OrderType: "MARKET",
		IsBuy:     true,
	})

	// so6 := order.New(order.NewStockOrderParams{
	// 	NewEntityParams: entity.NewEntityParams{
	// 		ID:           "so6",
	// 		DateCreated:  time.Now(),
	// 		DateModified: time.Now(),
	// 	},
	// 	StockID:   newStock1.GetId(),
	// 	Quantity:  10,
	// 	OrderType: "MARKET",
	// 	IsBuy:     true,
	// })
	//so2, so3, so4,, so6
	stockOrders := []order.StockOrderInterface{so1, so5}

	for _, so := range stockOrders {
		val, err = networkManager.OrderInitiator().Post("engine/placeStockOrder", so)
		if err != nil {
			panic(err)
		}
		println(string(val))
		//check prices
		val, err = networkManager.MatchingEngine().Get("transaction/getStockPrices", nil)
		if err != nil {
			panic(err)
		}
		println(string(val))
	}
	println("Stock Prices gotten")

	//cancel so3
	stockTransactionIdObject := network.StockTransactionID{StockTransactionID: so3.GetId()}
	val, err = networkManager.OrderInitiator().Post("engine/cancelStockTransaction", stockTransactionIdObject)
	if err != nil {
		panic(err)
	}
	println(string(val))

	// // fmt.Println the Stock Order
	// // fmt.Println("Stock Order: ")
	// // fmt.Println(so.GetId())
	// // fmt.Println(so.GetDateCreated())
	// // fmt.Println(so.GetDateModified())
	// // fmt.Println(so.GetStockID())
	// // fmt.Println(so.GetQuantity())
	// // fmt.Println(so.GetPrice())
	// // fmt.Println(so.GetOrderType())
	// // fmt.Println(so.GetIsBuy())

	// // //testing database.
	// // _databaseManager := databaseAccessStockOrder.NewDatabaseAccess(&databaseAccessStockOrder.NewDatabaseAccessParams{})

	// // fmt.Println("Database Test")
	// // fmt.Println("Testing create Stock Order: ")
	// // _databaseManager.Create(so)
	// // fmt.Println("Stock Order Created with ID: ", so.GetId())
	// // so2 := _databaseManager.GetByID(so.GetId())
	// // fmt.Println("Testing get Stock Order: ")
	// // fmt.Println(so2.GetId())
	// // fmt.Println(so2.GetDateCreated())
	// // fmt.Println(so2.GetDateModified())
	// // fmt.Println(so2.GetStockID())
	// // fmt.Println(so2.GetQuantity())
	// // fmt.Println(so2.GetPrice())
	// // fmt.Println(so2.GetOrderType())
	// // fmt.Println(so2.GetIsBuy())

	// // fmt.Println("Testing group get Stock Orders: ")
	// // idList := []string{"so1", so.GetId()}
	// // so5 := _databaseManager.GetByIDs(idList)
	// // for _, so6 := range *so5 {
	// // 	fmt.Println(so6.GetId())
	// // 	fmt.Println(so6.GetDateCreated())
	// // 	fmt.Println(so6.GetDateModified())
	// // 	fmt.Println(so6.GetStockID())
	// // 	fmt.Println(so6.GetQuantity())
	// // 	fmt.Println(so6.GetPrice())
	// // 	fmt.Println(so6.GetOrderType())
	// // 	fmt.Println(so6.GetIsBuy())
	// // }

	// // fmt.Println("Testing update Stock Order: ")
	// // so.SetIsBuy(false)
	// // _databaseManager.Update(so)
	// // so4 := _databaseManager.GetByID(so.GetId())
	// // fmt.Println("Stock Order: ")
	// // fmt.Println(so4.GetId())
	// // fmt.Println(so4.GetDateCreated())
	// // fmt.Println(so4.GetDateModified())
	// // fmt.Println(so4.GetStockID())
	// // fmt.Println(so4.GetQuantity())
	// // fmt.Println(so4.GetPrice())
	// // fmt.Println(so4.GetOrderType())
	// // fmt.Println(so4.GetIsBuy())

	// // fmt.Println("Testing delete Stock Order: ")
	// // _databaseManager.Delete(so.GetId())
	// // //_databaseManager.GetByID(so.GetId())

	// // fmt.Println("Database Test Complete")

	// // Create a new Stock Transaction
	// st1 := transaction.NewStockTransaction(transaction.NewStockTransactionParams{
	// 	NewEntityParams: entity.NewEntityParams{
	// 		ID:           "st1",
	// 		DateCreated:  time.Now(),
	// 		DateModified: time.Now(),
	// 	},
	// 	OrderStatus: "PENDING",
	// 	StockOrder:  so,
	// 	TimeStamp:   time.Now(),
	// })

	// // fmt.Println the Stock Transaction
	// //fmt.Println("Stock Transaction: ")
	// // fmt.Println(st1.GetId())
	// // fmt.Println(st1.GetDateCreated())
	// // fmt.Println(st1.GetDateModified())
	// // fmt.Println(st1.GetOrderStatus())
	// // fmt.Println(st1.GetStockID())
	// // fmt.Println(st1.GetParentStockTransactionID())
	// // fmt.Println(st1.GetWalletTransactionID())
	// // fmt.Println(st1.GetIsBuy())
	// // fmt.Println(st1.GetOrderType())
	// // fmt.Println(st1.GetStockPrice())
	// // fmt.Println(st1.GetQuantity())
	// // fmt.Println(st1.GetTimestamp())

	// //testing database.
	// network := network.NewNetwork()
	// _databaseManagerTransactions := databaseAccessTransaction.NewDatabaseAccess(&databaseAccessTransaction.NewDatabaseAccessParams{
	// 	Network: network,
	// })
	// _databaseManagerStockTransactions := _databaseManagerTransactions.StockTransaction()

	// fmt.Println("HTTP and Database Test")
	// fmt.Println("-----------------\nTesting create Stock Transaction: ")
	// st23, err := _databaseManagerStockTransactions.Create(st1)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println("Stock Transaction Created with ID: ", st23.GetId())
	// fmt.Println("-----------------")
	// // fmt.Println("Testing get Stock Transactions: ")
	// // st2, err := _databaseManagerStockTransactions.GetByID(st23.GetId())
	// // if err != nil {
	// // 	fmt.Println(err)
	// // }
	// // fmt.Println(st2.GetId())
	// // fmt.Println(st2.GetDateCreated())
	// // fmt.Println(st2.GetDateModified())
	// // fmt.Println(st2.GetOrderStatus())
	// // fmt.Println(st2.GetStockID())
	// // fmt.Println(st2.GetParentStockTransactionID())
	// // fmt.Println(st2.GetWalletTransactionID())
	// // fmt.Println(st2.GetIsBuy())
	// // fmt.Println(st2.GetOrderType())
	// // fmt.Println(st2.GetStockPrice())
	// // fmt.Println(st2.GetQuantity())
	// // fmt.Println(st2.GetTimestamp())
	// // fmt.Println("-----------------")
	// // fmt.Println("Testing group get Stock Transaction: ")
	// // idList := []string{"st1", st23.GetId()}
	// // st3, err := _databaseManagerStockTransactions.GetByIDs(idList)
	// // if err != nil {
	// // 	fmt.Println(err)
	// // }
	// // for _, st4 := range *st3 {
	// // 	fmt.Println(st4.GetId())
	// // 	fmt.Println(st4.GetDateCreated())
	// // 	fmt.Println(st4.GetDateModified())
	// // 	fmt.Println(st4.GetOrderStatus())
	// // 	fmt.Println(st4.GetStockID())
	// // 	fmt.Println(st4.GetParentStockTransactionID())
	// // 	fmt.Println(st4.GetWalletTransactionID())
	// // 	fmt.Println(st4.GetIsBuy())
	// // 	fmt.Println(st4.GetOrderType())
	// // 	fmt.Println(st4.GetStockPrice())
	// // 	fmt.Println(st4.GetQuantity())
	// // 	fmt.Println(st4.GetTimestamp())
	// // }
	// // fmt.Println("-----------------")
	// fmt.Println("Testing update Stock Transaction: ")
	// st23.SetOrderStatus("COMPLETE")
	// _databaseManagerStockTransactions.Update(st23)
	// st5, err := _databaseManagerStockTransactions.GetByID(st23.GetId())
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println("Stock Transaction: ")
	// fmt.Println(st5.GetId())
	// fmt.Println(st5.GetDateCreated())
	// fmt.Println(st5.GetDateModified())
	// fmt.Println(st5.GetOrderStatus())
	// fmt.Println(st5.GetStockID())
	// fmt.Println(st5.GetParentStockTransactionID())
	// fmt.Println(st5.GetWalletTransactionID())
	// fmt.Println(st5.GetIsBuy())
	// fmt.Println(st5.GetOrderType())
	// fmt.Println(st5.GetStockPrice())
	// fmt.Println(st5.GetQuantity())
	// fmt.Println(st5.GetTimestamp())
	// fmt.Println("-----------------")

	// fmt.Println("Testing group get Stock Transaction by foreign key: ", st23.GetStockID())
	// st6, err := _databaseManagerStockTransactions.GetByForeignID("stock_id", st23.GetStockID())
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// for _, st7 := range *st6 {
	// 	fmt.Println(st7.GetId())
	// 	fmt.Println(st7.GetDateCreated())
	// 	fmt.Println(st7.GetDateModified())
	// 	fmt.Println(st7.GetOrderStatus())
	// 	fmt.Println(st7.GetStockID())
	// 	fmt.Println(st7.GetParentStockTransactionID())
	// 	fmt.Println(st7.GetWalletTransactionID())
	// 	fmt.Println(st7.GetIsBuy())
	// 	fmt.Println(st7.GetOrderType())
	// 	fmt.Println(st7.GetStockPrice())
	// 	fmt.Println(st7.GetQuantity())
	// 	fmt.Println(st7.GetTimestamp())
	// }
	// fmt.Println("-----------------")

	// fmt.Println("Testing delete Stock Transaction: ")
	// err = _databaseManagerStockTransactions.Delete(st23.GetId())
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// _, err = _databaseManagerStockTransactions.GetByID(st23.GetId())
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println("-----------------")
	// fmt.Println("HTTP and Database Test Complete")
}
