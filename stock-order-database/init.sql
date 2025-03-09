CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE stockOrder (
    ID UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    Stock_ID UUID,
    Parent_Stock_Order_ID UUID,
    Date_Created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    Date_Modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    Is_Buy BOOLEAN NOT NULL,
    Order_Type TEXT NOT NULL,
    Price DECIMAL NOT NULL,
    Quantity INT NOT NULL,
    User_ID UUID NOT NULL,
    FOREIGN KEY (ParentStockOrderID) REFERENCES stockOrder(ID)
);