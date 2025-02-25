CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE Wallets (
    ID UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    User_ID UUID NOT NULL,
    Balance DECIMAL(18, 2) NOT NULL DEFAULT 0.00,
    DateCreated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    DateModified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE UserStocks (
    ID UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    UserID UUID NOT NULL,
    StockID UUID NOT NULL,
    StockName TEXT NOT NULL,
    Quantity INT NOT NULL,
    DateCreated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    DateModified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

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