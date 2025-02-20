CREATE TABLE stockTransactions (
    ID SERIAL PRIMARY KEY,
    StockID SERIAL,
    ParentStockTransactionID SERIAL,
    UserStockTransactionID SERIAL,
    WalletTransactionID SERIAL,
    DateCreated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    DateModified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    OrderStatus TEXT NOT NULL,
    IsBuy BOOLEAN NOT NULL,
    OrderType TEXT NOT NULL,
    StockPrice DECIMAL NOT NULL,
    Quantity INT NOT NULL,
    Timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (ParentStockTransactionID) REFERENCES stockTransactions(ID)
);

CREATE TABLE walletTransactions (
    ID SERIAL PRIMARY KEY,
    StockTransactionID SERIAL,
    WalletID SERIAL,
    UserStockTransactionID SERIAL,
    DateCreated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    DateModified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    IsDebit BOOLEAN NOT NULL,
    Amount DECIMAL NOT NULL,
    Timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (StockTransactionID) REFERENCES stockTransactions(ID)
);