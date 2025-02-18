CREATE TABLE stockTransactions (
    ID SERIAL PRIMARY KEY,
    StockID SERIAL,
    ParentStockTransactionID SERIAL,
    WalletTransactionID SERIAL,
    DateCreated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    DateModified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    OrderStatus TEXT NOT NULL,
    IsBuy BOOLEAN NOT NULL,
    OrderType TEXT NOT NULL,
    StockPrice DECIMAL NOT NULL,
    Quantity INT NOT NULL,
    FOREIGN KEY (ParentStockTransactionID) REFERENCES stockTransactions(ID)
);