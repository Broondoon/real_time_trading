package main

import (
	"Shared/entities/entity"
	"Shared/entities/order"
	"Shared/entities/stock"
	"Shared/entities/transaction"
	"Shared/entities/user"
	userStock "Shared/entities/user-stock"
	"Shared/entities/wallet"
	"log"
	"reflect"
	"time"
)

type Test[T any] interface {
	test(t T)
}

type TestStruct[T any] struct {
}

func (t *TestStruct[T]) test(td T) {

	log.Println("Test: %s", reflect.TypeOf(td))
}

func main() {

	// Create a new User
	u := user.New(user.NewUserParams{
		NewEntityParams: entity.NewEntityParams{
			ID:           "u1",
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		Username: "test",
		Password: "password",
	})

	test := TestStruct[entity.EntityInterface]{}
	test.test(u)

	// log.Println the User
	log.Println("User: ")
	log.Println(u.GetId())
	log.Println(u.GetDateCreated())
	log.Println(u.GetDateModified())
	log.Println(u.GetUsername())
	log.Println(u.GetPassword())
	log.Println(u.ToParams())
	log.Println(u.ToJSON())
	jsonData, err := u.ToJSON()
	if err != nil {
		log.Println("Error converting user to JSON:", err)
	} else {
		parsedUser, _ := user.Parse(jsonData)
		log.Println(parsedUser.GetId())
		log.Println(parsedUser.GetDateCreated())
		log.Println(parsedUser.GetDateModified())
		log.Println(parsedUser.GetUsername())
		log.Println(parsedUser.GetPassword())
	}

	// Create a new Stock
	s := stock.New(stock.NewStockParams{
		NewEntityParams: entity.NewEntityParams{
			ID:           "s1",
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		Name: "test",
	})

	// log.Println the Stock
	log.Println("Stock: ")
	log.Println(s.GetId())
	log.Println(s.GetDateCreated())
	log.Println(s.GetDateModified())
	log.Println(s.GetName())

	// Create a new Wallet
	w := wallet.New(wallet.NewWalletParams{
		NewEntityParams: entity.NewEntityParams{
			ID:           "w1",
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		User: u,
		// UserId:  "",
		Balance: 0.0,
	})
	// log.Println the Wallet
	log.Println("Wallet: ")
	log.Println(w.GetId())
	log.Println(w.GetDateCreated())
	log.Println(w.GetDateModified())
	log.Println(w.GetUserID())
	log.Println(w.GetBalance())
	log.Println(w.ToParams())
	log.Println(w.ToJSON())
	jsonData1, err := w.ToJSON()
	if err != nil {
		log.Println("Error converting user to JSON:", err)
	} else {
		parsedWallet, _ := wallet.Parse(jsonData1)
		log.Println(parsedWallet.GetId())
		log.Println(parsedWallet.GetDateCreated())
		log.Println(parsedWallet.GetDateModified())
		log.Println(parsedWallet.GetUserID())
		log.Println(parsedWallet.GetBalance())
	}

	//Create a new Stock Order
	so := order.New(order.NewStockOrderParams{
		NewEntityParams: entity.NewEntityParams{
			ID:           "so1",
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
	// log.Println the Stock Order
	log.Println("Stock Order: ")
	log.Println(so.GetId())
	log.Println(so.GetDateCreated())
	log.Println(so.GetDateModified())
	log.Println(so.GetStockID())
	log.Println(so.GetQuantity())
	log.Println(so.GetPrice())
	log.Println(so.GetOrderType())
	log.Println(so.GetIsBuy())

	// Create a new User Stock
	us := userStock.New(userStock.NewUserStockParams{
		NewEntityParams: entity.NewEntityParams{
			ID:           "us1",
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		User: u,
		// UserID: "",
		Stock: s,
		// StockID: "",
		Quantity: 0,
	})
	// log.Println the User Stock
	log.Println("User Stock: ")
	log.Println(us.GetId())
	log.Println(us.GetDateCreated())
	log.Println(us.GetDateModified())
	log.Println(us.GetUserID())
	log.Println(us.GetStockName())
	log.Println(us.GetQuantity())

	// Create a new Stock Transaction
	st1 := transaction.NewStockTransaction(transaction.NewStockTransactionParams{
		NewEntityParams: entity.NewEntityParams{
			ID:           "st1",
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		OrderStatus: "PENDING",
		StockOrder:  so,
	})

	// log.Println the Stock Transaction
	log.Println("Stock Transaction: ")
	log.Println(st1.GetId())
	log.Println(st1.GetDateCreated())
	log.Println(st1.GetDateModified())
	log.Println(st1.GetOrderStatus())
	log.Println(st1.GetStockID())
	log.Println(st1.GetParentStockTransactionID())
	log.Println(st1.GetWalletTransactionID())
	log.Println(st1.GetIsBuy())
	log.Println(st1.GetOrderType())
	log.Println(st1.GetStockPrice())
	log.Println(st1.GetQuantity())

	//Create a new wallet transaction
	wt := transaction.NewWalletTransaction(transaction.NewWalletTransactionParams{
		NewEntityParams: entity.NewEntityParams{
			ID:           "wt1",
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
	// log.Println the Wallet Transaction
	log.Println("Wallet Transaction: ")
	log.Println(wt.GetId())
	log.Println(wt.GetDateCreated())
	log.Println(wt.GetDateModified())
	log.Println(wt.GetWalletID())
	log.Println(wt.GetStockTransactionID())
	log.Println(wt.GetIsDebit())
	log.Println(wt.GetAmount())

	// Create a new Stock Transaction
	st2 := transaction.NewStockTransaction(transaction.NewStockTransactionParams{
		NewEntityParams: entity.NewEntityParams{
			ID:           "st2",
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		OrderStatus:            "COMPLETE",
		WalletTransaction:      wt,
		ParentStockTransaction: st1,
	})

	// log.Println the Stock Transaction
	log.Println("Stock Transaction: ")
	log.Println(st2.GetId())
	log.Println(st2.GetDateCreated())
	log.Println(st2.GetDateModified())
	log.Println(st2.GetOrderStatus())
	log.Println(st2.GetStockID())
	log.Println(st2.GetParentStockTransactionID())
	log.Println(st2.GetWalletTransactionID())
	log.Println(st2.GetIsBuy())
	log.Println(st2.GetOrderType())
	log.Println(st2.GetStockPrice())
	log.Println(st2.GetQuantity())
}
