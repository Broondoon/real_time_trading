CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE stockTransactions (
    ID UUID PRIMARY KEY,
    StockID UUID,
    ParentStockTransactionID UUID,
    UserStockTransactionID UUID,
    WalletTransactionID UUID,
    DateCreated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    DateModified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    OrderStatus TEXT NOT NULL,
    IsBuy BOOLEAN NOT NULL,
    OrderType TEXT NOT NULL,
    StockPrice DECIMAL NOT NULL,
    Quantity INT NOT NULL,
    UserID UUID NOT NULL,
    Timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (ParentStockTransactionID) REFERENCES stockTransactions(ID)
);

CREATE TABLE walletTransactions (
    ID UUID PRIMARY KEY,
    StockTransactionID UUID,
    WalletID UUID,
    UserStockTransactionID UUID,
    DateCreated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    DateModified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    IsDebit BOOLEAN NOT NULL,
    Amount DECIMAL NOT NULL,
    UserID UUID NOT NULL,
    Timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (StockTransactionID) REFERENCES stockTransactions(ID)
);