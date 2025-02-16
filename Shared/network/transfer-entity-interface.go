package network

type MatchingEngineToExectuionJSON struct {
	StockID     string  `json:"stock_id"`
	BuyOrderID  string  `json:"buy_order_id"`
	SellOrderID string  `json:"sell_order_id"`
	StockPrice  float64 `json:"stock_price"`
	Quantity    int     `json:"quantity"`
}
