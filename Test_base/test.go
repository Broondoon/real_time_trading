package main

import (
	"Shared/entities/entity"
	"Shared/entities/order"
	"Shared/entities/stock"
	"Shared/entities/transaction"
	"Shared/entities/user"
	userStock "Shared/entities/user-stock"
	"Shared/entities/wallet"
	"fmt"
	"time"
)

func main() {
	// Create a new User
	u := user.New(user.NewUserParams{
		NewEntityParams: entity.NewEntityParams{
			Id:           "u1",
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		UserProps: user.UserProps{
			Username: "test",
			Password: "password",
		},
	})

	// Print the User
	fmt.Print("User: ")
	fmt.Println(u.GetId())
	fmt.Println(u.GetDateCreated())
	fmt.Println(u.GetDateModified())
	fmt.Println(u.GetUsername())
	fmt.Println(u.GetPassword())
	fmt.Println(u.ToParams())
	fmt.Println(u.ToJSON())
	jsonData, err := u.ToJSON()
	if err != nil {
		fmt.Println("Error converting user to JSON:", err)
	} else {
		parsedUser, _ := user.Parse(jsonData)
		fmt.Println(parsedUser.GetId())
		fmt.Println(parsedUser.GetDateCreated())
		fmt.Println(parsedUser.GetDateModified())
		fmt.Println(parsedUser.GetUsername())
		fmt.Println(parsedUser.GetPassword())
	}

	// Create a new Stock
	s := stock.New(stock.NewStockParams{
		NewEntityParams: entity.NewEntityParams{
			Id:           "s1",
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		StockProps: stock.StockProps{
			Name: "test",
		},
	})

	// Print the Stock
	fmt.Print("Stock: ")
	fmt.Println(s.GetId())
	fmt.Println(s.GetDateCreated())
	fmt.Println(s.GetDateModified())
	fmt.Println(s.GetName())

	// Create a new Wallet
	w := wallet.New(wallet.NewWalletParams{
		NewEntityParams: entity.NewEntityParams{
			Id:           "w1",
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		User: u,
		WalletProps: wallet.WalletProps{
			// UserId:  "",
			Balance: 0.0,
		},
	})
	// Print the Wallet
	fmt.Print("Wallet: ")
	fmt.Println(w.GetId())
	fmt.Println(w.GetDateCreated())
	fmt.Println(w.GetDateModified())
	fmt.Println(w.GetUserID())
	fmt.Println(w.GetBalance())
	fmt.Println(w.ToParams())
	fmt.Println(w.ToJSON())
	jsonData1, err := w.ToJSON()
	if err != nil {
		fmt.Println("Error converting user to JSON:", err)
	} else {
		parsedWallet, _ := wallet.Parse(jsonData1)
		fmt.Println(parsedWallet.GetId())
		fmt.Println(parsedWallet.GetDateCreated())
		fmt.Println(parsedWallet.GetDateModified())
		fmt.Println(parsedWallet.GetUserID())
		fmt.Println(parsedWallet.GetBalance())
	}

	//Create a new Stock Order
	so := order.New(order.NewStockOrderParams{
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
	us := userStock.New(userStock.NewUserStockParams{
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
