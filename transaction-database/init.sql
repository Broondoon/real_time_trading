CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE stockTransactions (
    ID UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    Stock_ID UUID,
    Parent_StockTransaction_ID UUID,
    User_Stock_Transaction_ID UUID,
    Wallet_Transaction_ID UUID,
    Date_Created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    Date_Modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    Order_Status TEXT NOT NULL,
    Is_Buy BOOLEAN NOT NULL,
    Order_Type TEXT NOT NULL,
    Stock_Price DECIMAL NOT NULL,
    Quantity INT NOT NULL,
    User_ID UUID NOT NULL,
    Timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (ParentStockTransactionID) REFERENCES stockTransactions(ID)
);

CREATE TABLE walletTransactions (
    ID UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    Stock_Transaction_ID UUID,
    Wallet_ID UUID,
    User_Stock_Transaction_ID UUID,
    Date_Created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    Date_Modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    Is_Debit BOOLEAN NOT NULL,
    Amount DECIMAL NOT NULL,
    User_ID UUID NOT NULL,
    Timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (StockTransactionID) REFERENCES stockTransactions(ID)
);
