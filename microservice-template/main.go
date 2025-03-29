package main

import (
	"Shared/entities/user"
	networkHttp "Shared/network/http"
	"log"
	//"Shared/network"
)

func main() {
	var err error
	//var val []byte

	networkManager := networkHttp.NewNetworkHttp()
	testUser := user.User{
		Username: "VanguardETF",
		Password: "Vang@123",
		Name:     "Vanguard Corp.",
	}

	test, err := networkManager.Authentication().Post("authentication/register", testUser)
	if err != nil {
		panic(err)
	}
	log.Println(string(test))
	test, err = networkManager.Authentication().Post("authentication/login", testUser)
	if err != nil {
		panic(err)
	}
	log.Println(string(test))
	// // {"user_name":"VanguardETF", "password":"Vang@123", "name":"Vanguard Corp."}
	// // val, err = networkManager.Transactions().Get("transaction/getStockTransactions", nil)
	// // if err != nil {
	// // 	panic(err)
	// // }
	// // log.Println(string(val))
	// // log.Println("Stock Transactions gotten")

	// ea := databaseAccessUserManagement.NewDatabaseAccess(&databaseAccessUserManagement.NewDatabaseAccessParams{
	// 	Network: networkManager,
	// })

	// // test, err := ea.Wallet().Create(wallet.New(wallet.NewWalletParams{
	// // 	NewEntityParams: entity.NewEntityParams{
	// // 		DateCreated:  time.Now(),
	// // 		DateModified: time.Now(),
	// // 	},
	// // 	UserID:  "6fd2fc6b-9142-4777-8b30-575ff6fa2460",
	// // 	Balance: 100000,
	// // }))
	// // if err != nil {
	// // 	log.Println("Error creating wallet: ", err)
	// // 	panic(err)
	// // }
	// // walletOutput, err := test.ToJSON()
	// // log.Println("Wallet created: ", string(walletOutput))

	// // testArray, err := ea.Wallet().GetByForeignID("UserID", "6fd2fc6b-9142-4777-8b30-575ff6fa2460")
	// // if err != nil {
	// // 	log.Println("Error getting wallet: ", err)
	// // 	panic(err)
	// // }
	// // log.Println(testArray)
	// // log.Println("Wallet gotten")

	// //Create a new Stock
	// newStock1 := stock.New(stock.NewStockParams{
	// 	NewEntityParams: entity.NewEntityParams{
	// 		ID:           uuid.New().String(),
	// 		DateCreated:  time.Now(),
	// 		DateModified: time.Now(),
	// 	},
	// 	Name: "Apple",
	// })
	// newStock2 := stock.New(stock.NewStockParams{
	// 	NewEntityParams: entity.NewEntityParams{
	// 		ID:           uuid.New().String(),
	// 		DateCreated:  time.Now(),
	// 		DateModified: time.Now(),
	// 	},
	// 	Name: "google",
	// })

	// // val, err = networkManager.Stocks().Post("setup/createStock", newStock1)
	// // if err != nil {
	// // 	panic(err)
	// // }

	// // val, err = networkManager.Stocks().Post("setup/createStock", newStock2)
	// // if err != nil {
	// // 	panic(err)
	// // }
	// // log.Println("Stock Created")

	// ea.UserStock().Create(userStock.New(userStock.NewUserStockParams{
	// 	NewEntityParams: entity.NewEntityParams{
	// 		ID: uuid.New().String(),
	// 	},
	// 	UserID:   "6fd2fc6b-9142-4777-8b30-575ff6fa2460",
	// 	StockID:  newStock1.GetId(),
	// 	Quantity: 100,
	// }))

	// ea.UserStock().Create(userStock.New(userStock.NewUserStockParams{
	// 	NewEntityParams: entity.NewEntityParams{
	// 		ID: uuid.New().String(),
	// 	},
	// 	UserID:   "6fd2fc6b-9142-4777-8b30-575ff6fa2460",
	// 	StockID:  newStock2.GetId(),
	// 	Quantity: 100,
	// }))

	// d, err := ea.UserStock().GetByForeignID("UserID", "6fd2fc6b-9142-4777-8b30-575ff6fa2460")
	// if err != nil {
	// 	log.Println("Error getting user stock: ", err)
	// 	panic(err)
	// }
	// for _, us := range *d {
	// 	da, err := us.ToJSON()
	// 	if err != nil {
	// 		log.Println("Error converting user stock to json: ", err)
	// 		panic(err)
	// 	}
	// 	log.Println("User Stock gotten: ", string(da))
	// 	us.UpdateQuantity(60)
	// 	ea.UserStock().Update(us)
	// }

	// // //Create a new Stock Order
	// // so1 := order.New(order.NewStockOrderParams{
	// // 	NewEntityParams: entity.NewEntityParams{
	// // 		ID:           "so1",
	// // 		DateCreated:  time.Now(),
	// // 		DateModified: time.Now(),
	// // 	},
	// // 	StockID:   newStock1.GetId(),
	// // 	Quantity:  4,
	// // 	Price:     7.5,
	// // 	OrderType: "LIMIT",
	// // 	IsBuy:     false,
	// // })

	// // // so2 := order.New(order.NewStockOrderParams{
	// // // 	NewEntityParams: entity.NewEntityParams{
	// // // 		ID:           "so2",
	// // // 		DateCreated:  time.Now(),
	// // // 		DateModified: time.Now(),
	// // // 	},
	// // // 	StockID:   newStock1.GetId(),
	// // // 	Quantity:  5,
	// // // 	Price:     7.5,
	// // // 	OrderType: "LIMIT",
	// // // 	IsBuy:     false,
	// // // })

	// // // so3 := order.New(order.NewStockOrderParams{

	// // // 	NewEntityParams: entity.NewEntityParams{
	// // // 		ID:           "so3",
	// // // 		DateCreated:  time.Now(),
	// // // 		DateModified: time.Now(),
	// // // 	},
	// // // 	StockID:   newStock1.GetId(),
	// // // 	Quantity:  5,
	// // // 	Price:     8.5,
	// // // 	OrderType: "LIMIT",
	// // // 	IsBuy:     false,
	// // // })

	// // so4 := order.New(order.NewStockOrderParams{
	// // 	NewEntityParams: entity.NewEntityParams{
	// // 		ID:           "so4",
	// // 		DateCreated:  time.Now(),
	// // 		DateModified: time.Now(),
	// // 	},
	// // 	StockID:   newStock1.GetId(),
	// // 	Quantity:  5,
	// // 	Price:     6.5,
	// // 	OrderType: "LIMIT",
	// // 	IsBuy:     false,
	// // })

	// so5 := order.New(order.NewStockOrderParams{
	// 	NewEntityParams: entity.NewEntityParams{
	// 		ID:           "so5",
	// 		DateCreated:  time.Now(),
	// 		DateModified: time.Now(),
	// 	},
	// 	StockID:   newStock1.GetId(),
	// 	Quantity:  5,
	// 	OrderType: "MARKET",
	// 	IsBuy:     true,
	// })

	// // so6 := order.New(order.NewStockOrderParams{
	// // 	NewEntityParams: entity.NewEntityParams{
	// // 		ID:           "so6",
	// // 		DateCreated:  time.Now(),
	// // 		DateModified: time.Now(),
	// // 	},
	// // 	StockID:   newStock1.GetId(),
	// // 	Quantity:  10,
	// // 	OrderType: "MARKET",
	// // 	IsBuy:     true,
	// // })
	// //so2, so3, so4,, so6
	// stockOrders := []order.StockOrderInterface{so1, so5}

	// for _, so := range stockOrders {
	// 	val, err = networkManager.OrderInitiator().Post("engine/placeStockOrder", so)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	log.Println(string(val))
	// 	//check prices
	// 	val, err = networkManager.MatchingEngine().Get("transaction/getStockPrices", nil)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	log.Println(string(val))
	// }
	// log.Println("Stock Prices gotten")

	// // //cancel so3
	// // stockTransactionIdObject := network.StockTransactionID{StockTransactionID: so3.GetId()}
	// // val, err = networkManager.OrderInitiator().Post("engine/cancelStockTransaction", stockTransactionIdObject)
	// // if err != nil {
	// // 	panic(err)
	// // }
	// // log.Println(string(val))

	// // // log.Println the Stock Order
	// // // log.Println("Stock Order: ")
	// // // log.Println(so.GetId())
	// // // log.Println(so.GetDateCreated())
	// // // log.Println(so.GetDateModified())
	// // // log.Println(so.GetStockID())
	// // // log.Println(so.GetQuantity())
	// // // log.Println(so.GetPrice())
	// // // log.Println(so.GetOrderType())
	// // // log.Println(so.GetIsBuy())

	// // // //testing database.
	// // // _databaseManager := databaseAccessStockOrder.NewDatabaseAccess(&databaseAccessStockOrder.NewDatabaseAccessParams{})

	// // // log.Println("Database Test")
	// // // log.Println("Testing create Stock Order: ")
	// // // _databaseManager.Create(so)
	// // // log.Println("Stock Order Created with ID: ", so.GetId())
	// // // so2 := _databaseManager.GetByID(so.GetId())
	// // // log.Println("Testing get Stock Order: ")
	// // // log.Println(so2.GetId())
	// // // log.Println(so2.GetDateCreated())
	// // // log.Println(so2.GetDateModified())
	// // // log.Println(so2.GetStockID())
	// // // log.Println(so2.GetQuantity())
	// // // log.Println(so2.GetPrice())
	// // // log.Println(so2.GetOrderType())
	// // // log.Println(so2.GetIsBuy())

	// // // log.Println("Testing group get Stock Orders: ")
	// // // idList := []string{"so1", so.GetId()}
	// // // so5 := _databaseManager.GetByIDs(idList)
	// // // for _, so6 := range *so5 {
	// // // 	log.Println(so6.GetId())
	// // // 	log.Println(so6.GetDateCreated())
	// // // 	log.Println(so6.GetDateModified())
	// // // 	log.Println(so6.GetStockID())
	// // // 	log.Println(so6.GetQuantity())
	// // // 	log.Println(so6.GetPrice())
	// // // 	log.Println(so6.GetOrderType())
	// // // 	log.Println(so6.GetIsBuy())
	// // // }

	// // // log.Println("Testing update Stock Order: ")
	// // // so.SetIsBuy(false)
	// // // _databaseManager.Update(so)
	// // // so4 := _databaseManager.GetByID(so.GetId())
	// // // log.Println("Stock Order: ")
	// // // log.Println(so4.GetId())
	// // // log.Println(so4.GetDateCreated())
	// // // log.Println(so4.GetDateModified())
	// // // log.Println(so4.GetStockID())
	// // // log.Println(so4.GetQuantity())
	// // // log.Println(so4.GetPrice())
	// // // log.Println(so4.GetOrderType())
	// // // log.Println(so4.GetIsBuy())

	// // // log.Println("Testing delete Stock Order: ")
	// // // _databaseManager.Delete(so.GetId())
	// // // //_databaseManager.GetByID(so.GetId())

	// // // log.Println("Database Test Complete")

	// // // Create a new Stock Transaction
	// // st1 := transaction.NewStockTransaction(transaction.NewStockTransactionParams{
	// // 	NewEntityParams: entity.NewEntityParams{
	// // 		ID:           "st1",
	// // 		DateCreated:  time.Now(),
	// // 		DateModified: time.Now(),
	// // 	},
	// // 	OrderStatus: "PENDING",
	// // 	StockOrder:  so,
	// // 	TimeStamp:   time.Now(),
	// // })

	// // // log.Println the Stock Transaction
	// // //log.Println("Stock Transaction: ")
	// // // log.Println(st1.GetId())
	// // // log.Println(st1.GetDateCreated())
	// // // log.Println(st1.GetDateModified())
	// // // log.Println(st1.GetOrderStatus())
	// // // log.Println(st1.GetStockID())
	// // // log.Println(st1.GetParentStockTransactionID())
	// // // log.Println(st1.GetWalletTransactionID())
	// // // log.Println(st1.GetIsBuy())
	// // // log.Println(st1.GetOrderType())
	// // // log.Println(st1.GetStockPrice())
	// // // log.Println(st1.GetQuantity())
	// // // log.Println(st1.GetTimestamp())

	// // //testing database.
	// // network := network.NewNetwork()
	// // _databaseManagerTransactions := databaseAccessTransaction.NewDatabaseAccess(&databaseAccessTransaction.NewDatabaseAccessParams{
	// // 	Network: network,
	// // })
	// // _databaseManagerStockTransactions := _databaseManagerTransactions.StockTransaction()

	// // log.Println("HTTP and Database Test")
	// // log.Println("-----------------\nTesting create Stock Transaction: ")
	// // st23, err := _databaseManagerStockTransactions.Create(st1)
	// // if err != nil {
	// // 	log.Println(err)
	// // }
	// // log.Println("Stock Transaction Created with ID: ", st23.GetId())
	// // log.Println("-----------------")
	// // // log.Println("Testing get Stock Transactions: ")
	// // // st2, err := _databaseManagerStockTransactions.GetByID(st23.GetId())
	// // // if err != nil {
	// // // 	log.Println(err)
	// // // }
	// // // log.Println(st2.GetId())
	// // // log.Println(st2.GetDateCreated())
	// // // log.Println(st2.GetDateModified())
	// // // log.Println(st2.GetOrderStatus())
	// // // log.Println(st2.GetStockID())
	// // // log.Println(st2.GetParentStockTransactionID())
	// // // log.Println(st2.GetWalletTransactionID())
	// // // log.Println(st2.GetIsBuy())
	// // // log.Println(st2.GetOrderType())
	// // // log.Println(st2.GetStockPrice())
	// // // log.Println(st2.GetQuantity())
	// // // log.Println(st2.GetTimestamp())
	// // // log.Println("-----------------")
	// // // log.Println("Testing group get Stock Transaction: ")
	// // // idList := []string{"st1", st23.GetId()}
	// // // st3, err := _databaseManagerStockTransactions.GetByIDs(idList)
	// // // if err != nil {
	// // // 	log.Println(err)
	// // // }
	// // // for _, st4 := range *st3 {
	// // // 	log.Println(st4.GetId())
	// // // 	log.Println(st4.GetDateCreated())
	// // // 	log.Println(st4.GetDateModified())
	// // // 	log.Println(st4.GetOrderStatus())
	// // // 	log.Println(st4.GetStockID())
	// // // 	log.Println(st4.GetParentStockTransactionID())
	// // // 	log.Println(st4.GetWalletTransactionID())
	// // // 	log.Println(st4.GetIsBuy())
	// // // 	log.Println(st4.GetOrderType())
	// // // 	log.Println(st4.GetStockPrice())
	// // // 	log.Println(st4.GetQuantity())
	// // // 	log.Println(st4.GetTimestamp())
	// // // }
	// // // log.Println("-----------------")
	// // log.Println("Testing update Stock Transaction: ")
	// // st23.SetOrderStatus("COMPLETE")
	// // _databaseManagerStockTransactions.Update(st23)
	// // st5, err := _databaseManagerStockTransactions.GetByID(st23.GetId())
	// // if err != nil {
	// // 	log.Println(err)
	// // }
	// // log.Println("Stock Transaction: ")
	// // log.Println(st5.GetId())
	// // log.Println(st5.GetDateCreated())
	// // log.Println(st5.GetDateModified())
	// // log.Println(st5.GetOrderStatus())
	// // log.Println(st5.GetStockID())
	// // log.Println(st5.GetParentStockTransactionID())
	// // log.Println(st5.GetWalletTransactionID())
	// // log.Println(st5.GetIsBuy())
	// // log.Println(st5.GetOrderType())
	// // log.Println(st5.GetStockPrice())
	// // log.Println(st5.GetQuantity())
	// // log.Println(st5.GetTimestamp())
	// // log.Println("-----------------")

	// // log.Println("Testing group get Stock Transaction by foreign key: ", st23.GetStockID())
	// // st6, err := _databaseManagerStockTransactions.GetByForeignID("stock_id", st23.GetStockID())
	// // if err != nil {
	// // 	log.Println(err)
	// // }
	// // for _, st7 := range *st6 {
	// // 	log.Println(st7.GetId())
	// // 	log.Println(st7.GetDateCreated())
	// // 	log.Println(st7.GetDateModified())
	// // 	log.Println(st7.GetOrderStatus())
	// // 	log.Println(st7.GetStockID())
	// // 	log.Println(st7.GetParentStockTransactionID())
	// // 	log.Println(st7.GetWalletTransactionID())
	// // 	log.Println(st7.GetIsBuy())
	// // 	log.Println(st7.GetOrderType())
	// // 	log.Println(st7.GetStockPrice())
	// // 	log.Println(st7.GetQuantity())
	// // 	log.Println(st7.GetTimestamp())
	// // }
	// // log.Println("-----------------")

	// // log.Println("Testing delete Stock Transaction: ")
	// // err = _databaseManagerStockTransactions.Delete(st23.GetId())
	// // if err != nil {
	// // 	log.Println(err)
	// // }
	// // _, err = _databaseManagerStockTransactions.GetByID(st23.GetId())
	// // if err != nil {
	// // 	log.Println(err)
	// // }
	// // log.Println("-----------------")
	// // log.Println("HTTP and Database Test Complete")
}
