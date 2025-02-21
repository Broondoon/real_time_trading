package network

type MatchingEngineToExecutionJSON struct {
	StockID       string  `json:"stock_id"`
	BuyOrderID    string  `json:"buy_order_id"`
	SellOrderID   string  `json:"sell_order_id"`
	IsBuyPartial  bool    `json:"is_buy_partial"`
	IsSellPartial bool    `json:"is_sell_partial"`
	StockPrice    float64 `json:"stock_price"`
	Quantity      int     `json:"quantity"`
}

type StockPrices struct {
	StockPrices map[string]float64 `json:"stock_prices"`
}

type StockID struct {
	StockID string `json:"stock_id"`
}

type StockTransactionID struct {
	StockTransactionID string `json:"stock_tx_id"`
}
