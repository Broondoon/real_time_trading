package orderExecutorService

import (
	//"Shared/entities/entity"
	"Shared/entities/transaction"
	userStock "Shared/entities/user-stock"
	"databaseAccessTransaction"
	"databaseAccessUserManagement"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// Calculates the total cost of a transaction given the quantity and stock price.
func calculateTotalTransactionCost(quantity int, stockPrice float64) float64 {
	return float64(quantity) * stockPrice
}

// Finds and validates user stock portfolios
func findUserStockPortfolios(
	buyerID *uuid.UUID,
	sellerID *uuid.UUID,
	databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface,
) (*[]userStock.UserStockInterface, *[]userStock.UserStockInterface, error) {

	// Get buyer's current stock holdings
	buyerStockPortfolio, err := databaseAccessUser.UserStock().GetUserStocks(buyerID.String())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get buyer stocks: %v", err)
	}
	log.Printf("Retrieved buyer portfolio with %d stocks", len(*buyerStockPortfolio))

	// Get seller's current stock holdings
	sellerStockPortfolio, err := databaseAccessUser.UserStock().GetUserStocks(sellerID.String())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get seller stocks: %v", err)
	}
	log.Printf("%s", fmt.Sprintf("Retrieved seller portfolio with %d stocks", len(*sellerStockPortfolio)))

	return buyerStockPortfolio, sellerStockPortfolio, nil
}

// Finds and validates seller's stock holding
func handleSellerStock(
	sellerStockPortfolio *[]userStock.UserStockInterface,
	stockID *uuid.UUID,
	quantity int,
) (userStock.UserStockInterface, error) {

	var sellerStock userStock.UserStockInterface
	for _, stock := range *sellerStockPortfolio {
		if stock.GetStockIDString() == stockID.String() {
			sellerStock = stock
			break
		}
	}

	return sellerStock, nil
}

// Creates or retrieves buyer's stock holding
func handleBuyerStock(
	buyerStockPortfolio *[]userStock.UserStockInterface,
	buyerID *uuid.UUID,
	stockID *uuid.UUID,
	//quantity int,
	sellerStock userStock.UserStockInterface,
	databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface,
) (userStock.UserStockInterface, error) {

	var buyerStock userStock.UserStockInterface
	for _, stock := range *buyerStockPortfolio {
		if stock.GetStockIDString() == stockID.String() {
			buyerStock = stock
			break
		}
	}
	log.Printf("%s", fmt.Sprintf("Initially Buyer has %d shares of StockID: %s", buyerStock.GetQuantity(), buyerStock.GetStockID()))

	// If the buyer doesn't have any of the stock, a new stock holding is created
	// The quantity is originally set to zero and is updated after (otherwise there's an error where the quantity is double what it should be)

	if buyerStock == nil {
		buyerStock = userStock.New(userStock.NewUserStockParams{
			UserID:    buyerID,
			StockID:   stockID,
			StockName: sellerStock.GetStockName(),
			Quantity:  0,
		})
		createdStock, err := databaseAccessUser.UserStock().Create(buyerStock)
		if err != nil {
			return nil, fmt.Errorf("failed to create buyer stock holding: %v", err)
		}
		buyerStock = createdStock
	}

	return buyerStock, nil
}

// Updates the user's stock quantities in the database
func updateUserStockQuantities(
	buyerStock userStock.UserStockInterface,
	sellerStock userStock.UserStockInterface,
	quantity int,
	databaseAccessUser databaseAccessUserManagement.DatabaseAccessInterface,
) error {
	log.Println("Updating user stock quantities")
	log.Println(fmt.Sprintf("Initial -> Seller has %d shares of StockID: %s.", sellerStock.GetQuantity(), sellerStock.GetStockID()), "\n", fmt.Sprintf("Initial -> Buyer has %d shares of StockID: %s", buyerStock.GetQuantity(), buyerStock.GetStockID()), "\nQuantity: ", quantity)
	buyerJson, err := buyerStock.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to convert buyer stock to JSON: %v", err)
	}
	log.Println("Buyer Stock: ", string(buyerJson))
	sellerJson, err := sellerStock.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to convert seller stock to JSON: %v", err)
	}
	log.Println("Seller Stock: ", string(sellerJson))

	buyerStock.UpdateQuantity(quantity)

	if err := databaseAccessUser.UserStock().Update(buyerStock); err != nil {
		return fmt.Errorf("failed to update buyer stock: %v", err)
	}
	//log.Println(fmt.Sprintf("Final -> Buyer  has %d shares of StockID: %s", buyerStock.GetQuantity(), buyerStock.GetStockID()))

	log.Printf("Final Buyer Quantity of StockID = %s is %d, Final Seller Quantity of StockID = %s is %d", buyerStock.GetStockID(), buyerStock.GetQuantity(), sellerStock.GetStockID(), sellerStock.GetQuantity())

	return nil
}

// Updates transaction status and creates filled transaction if needed
func updateTransactionStatus(
	stockTx transaction.StockTransactionInterface,
	isBuyPartial bool,
	isSellPartial bool,
	stockPrice float64,
	databaseAccessTransact databaseAccessTransaction.DatabaseAccessInterface,
) error {

	log.Printf("%s", fmt.Sprintf("BEFORE Update Status: %s", stockTx.GetOrderStatus()))

	// Set the stock price in the transaction

	// Handle partial matching for both buy and sell orders
	if stockTx.GetIsBuy() {
		stockTx.UpdateStockPrice(stockPrice)
		if isBuyPartial {
			stockTx.SetOrderStatus("PARTIALLY_COMPLETE")
		} else {
			stockTx.SetOrderStatus("COMPLETED")
		}

	} else {
		// For sell orders
		if isSellPartial {
			stockTx.SetOrderStatus("PARTIALLY_COMPLETE")
		} else {
			stockTx.SetOrderStatus("COMPLETED")
		}
	}

	// Update in database
	if err := databaseAccessTransact.StockTransaction().Update(stockTx); err != nil {
		return fmt.Errorf("failed to update transaction status: %v", err)
	}
	log.Printf("%s", fmt.Sprintf("AFTER Update Status: %s", stockTx.GetOrderStatus()))

	// Create filled transaction for partial orders
	if isBuyPartial || isSellPartial {
		filledTx := transaction.NewStockTransaction(transaction.NewStockTransactionParams{
			ParentStockTransaction: stockTx,
			OrderStatus:            "COMPLETED", // Child transaction is always COMPLETED
			TimeStamp:              time.Now(),
			StockPrice:             stockPrice,
		})

		if _, err := databaseAccessTransact.StockTransaction().Create(filledTx); err != nil {
			return fmt.Errorf("failed to create filled stock transaction: %v", err)
		}

		log.Printf("%s", fmt.Sprintf("Created Filled Transaction with ID: %s", filledTx.GetId()))
	}

	return nil
}
