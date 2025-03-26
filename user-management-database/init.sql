CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE Wallets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    balance DECIMAL(18, 2) NOT NULL DEFAULT 0.00,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE UserStocks (
    id UUID DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    stock_id UUID NOT NULL,
    stock_name TEXT NOT NULL,
    quantity INT NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    PRIMARY KEY (user_id, stock_id)
);

CREATE INDEX idx_user_stocks_user_id_stock_id
    ON UserStocks (user_id, stock_id);

/*
INSERT INTO Wallets (ID, UserID, Balance)
VALUES (uuid_generate_v4(), '6fd2fc6b-9142-4777-8b30-575ff6fa2460', 1000.00);

INSERT INTO UserStocks (ID, UserID, StockID, StockName, Quantity)
VALUES (uuid_generate_v4(), '6fd2fc6b-9142-4777-8b30-575ff6fa2460', 1, 'AAPL', 50);

INSERT INTO UserStocks (ID, UserID, StockID, StockName, Quantity)
VALUES (uuid_generate_v4(), '6fd2fc6b-9142-4777-8b30-575ff6fa2460', 2, 'GOOGL', 30);

INSERT INTO UserStocks (ID, UserID, StockID, StockName, Quantity)
VALUES (uuid_generate_v4(), '6fd2fc6b-9142-4777-8b30-575ff6fa2460', 3, 'MSFT', 40);
*/