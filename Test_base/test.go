package main

import (
	"Shared/entities/entity"
	"Shared/entities/order"
	"Shared/entities/stock"
	"Shared/entities/transaction"
	"Shared/entities/user"
	"Shared/entities/wallet"
	"fmt"
	"time"
)

func main() {
	// Create a new User
	u := user.NewUser(user.NewUserParams{
		NewEntityParams: entity.NewEntityParams{
			Id:           "u1",
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		Username: "test",
		Password: "password",
	})

	// Print the User
	fmt.Print("User: ")
	fmt.Println(u.GetId())
	fmt.Println(u.GetDateCreated())
	fmt.Println(u.GetDateModified())
	fmt.Println(u.GetUsername())
	fmt.Println(u.GetPassword())

	// Create a new Stock
	s := stock.NewStock(stock.NewStockParams{
		NewEntityParams: entity.NewEntityParams{
			Id:           "s1",
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		Name: "test",
	})

	// Print the Stock
	fmt.Print("Stock: ")
	fmt.Println(s.GetId())
	fmt.Println(s.GetDateCreated())
	fmt.Println(s.GetDateModified())
	fmt.Println(s.GetName())

	// Create a new Wallet
	w := wallet.NewWallet(wallet.NewWalletParams{
		NewEntityParams: entity.NewEntityParams{
			Id:           "w1",
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		User: u,
		// UserId:  "",
		Balance: 0.0,
	})
	// Print the Wallet
	fmt.Print("Wallet: ")
	fmt.Println(w.GetId())
	fmt.Println(w.GetDateCreated())
	fmt.Println(w.GetDateModified())
	fmt.Println(w.GetUserID())
	fmt.Println(w.GetBalance())

	//Create a new Stock Order
	so := order.NewStockOrder(order.NewStockOrderParams{
		NewEntityParams: entity.NewEntityParams{
			Id:           "so1",
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		Stock: s,
		// StockId: "",
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

	// Create a new User Stock
	us := stock.NewUserStock(stock.NewUserStockParams{
		NewEntityParams: entity.NewEntityParams{
			Id:           "us1",
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		User: u,
		// UserID: "",
		Stock: s,
		// StockID: "",
		Quantity: 0,
	})
	// Print the User Stock
	fmt.Print("User Stock: ")
	fmt.Println(us.GetId())
	fmt.Println(us.GetDateCreated())
	fmt.Println(us.GetDateModified())
	fmt.Println(us.GetUserID())
	fmt.Println(us.GetStockName())
	fmt.Println(us.GetQuantity())

	// Create a new Stock Transaction
	st1 := transaction.NewStockTransaction(transaction.NewStockTransactionParams{
		NewEntityParams: entity.NewEntityParams{
			Id:           "st1",
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		OrderStatus: "PENDING",
		StockOrder:  so,
	})

	// Print the Stock Transaction
	fmt.Print("Stock Transaction: ")
	fmt.Println(st1.GetId())
	fmt.Println(st1.GetDateCreated())
	fmt.Println(st1.GetDateModified())
	fmt.Println(st1.GetOrderStatus())
	fmt.Println(st1.GetStockID())
	fmt.Println(st1.GetParentStockTransactionID())
	fmt.Println(st1.GetWalletTransactionID())
	fmt.Println(st1.GetIsBuy())
	fmt.Println(st1.GetOrderType())
	fmt.Println(st1.GetStockPrice())
	fmt.Println(st1.GetQuantity())

	//Create a new wallet transaction
	wt := transaction.NewWalletTransaction(transaction.NewWalletTransactionParams{
		NewEntityParams: entity.NewEntityParams{
			Id:           "wt1",
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		Wallet: w,
		// WalletID: "",
		StockTransaction: st1,
		// StockTransactionID: "",
		IsDebit: true,
		Amount:  0.0,
	})
	// Print the Wallet Transaction
	fmt.Print("Wallet Transaction: ")
	fmt.Println(wt.GetId())
	fmt.Println(wt.GetDateCreated())
	fmt.Println(wt.GetDateModified())
	fmt.Println(wt.GetWalletID())
	fmt.Println(wt.GetStockTransactionID())
	fmt.Println(wt.GetIsDebit())
	fmt.Println(wt.GetAmount())

	// Create a new Stock Transaction
	st2 := transaction.NewStockTransaction(transaction.NewStockTransactionParams{
		NewEntityParams: entity.NewEntityParams{
			Id:           "st2",
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		OrderStatus:            "COMPLETE",
		WalletTransaction:      wt,
		ParentStockTransaction: st1,
	})

	// Print the Stock Transaction
	fmt.Print("Stock Transaction: ")
	fmt.Println(st2.GetId())
	fmt.Println(st2.GetDateCreated())
	fmt.Println(st2.GetDateModified())
	fmt.Println(st2.GetOrderStatus())
	fmt.Println(st2.GetStockID())
	fmt.Println(st2.GetParentStockTransactionID())
	fmt.Println(st2.GetWalletTransactionID())
	fmt.Println(st2.GetIsBuy())
	fmt.Println(st2.GetOrderType())
	fmt.Println(st2.GetStockPrice())
	fmt.Println(st2.GetQuantity())
}
