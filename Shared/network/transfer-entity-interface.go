package network

type MatchingEngineToExecutionJSON struct {
	BuyerID       string  `json:"buyer_id"`
	SellerID      string  `json:"seller_id"`
	StockID       string  `json:"stock_id"`
	BuyOrderID    string  `json:"buy_order_id"`
	SellOrderID   string  `json:"sell_order_id"`
	IsBuyPartial  bool    `json:"is_buy_partial"`
	IsSellPartial bool    `json:"is_sell_partial"`
	StockPrice    float64 `json:"stock_price"`
	Quantity      int     `json:"quantity"`
}

type StockPrice struct {
	StockID   string  `json:"stock_id"`
	StockName string  `json:"stock_name"`
	Price     float64 `json:"current_price"`
}

type StockID struct {
	StockID string `json:"stock_id"`
}

type StockTransactionID struct {
	StockTransactionID string `json:"stock_tx_id"`
}

type WalletBalance struct {
	Balance float64 `json:"balance"`
}

type AddStock struct {
	StockID  string `json:"stock_id"`
	Quantity int    `json:"quantity"`
}

// Transfer Entity to send back to Matching Engine

// If the buy order failed, then the is_buy_failed field = true
// If the sell order failed, then the is_sell_failed field =true
type ExecutorToMatchingEngineJSON struct {
	IsBuyFailure  bool `json:"is_buy_failed"`
	IsSellFailure bool `json:"is_sell_failed"`
}
