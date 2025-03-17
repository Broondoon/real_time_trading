CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE stockTransactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stock_id UUID,
    parent_stock_transaction_id UUID,
    user_stock_transaction_id UUID,
    wallet_transaction_id UUID,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    order_status TEXT NOT NULL,
    is_buy BOOLEAN NOT NULL,
    order_type TEXT NOT NULL,
    stock_price DECIMAL NOT NULL,
    quantity INT NOT NULL,
    user_id UUID NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_stock_transaction_id) REFERENCES stock_transactions(id)
);

CREATE TABLE walletTransactions (
    id uuid primary key default uuid_generate_v4(),
    stock_transaction_id uuid,
    wallet_id uuid,
    user_stock_transaction_id uuid,
    date_created timestamp default current_timestamp,
    date_modified timestamp default current_timestamp,
    is_debit boolean not null,
    amount decimal not null,
    user_id uuid not null,
    timestamp timestamp default current_timestamp,
    foreign key (stock_transaction_id) references stock_transactions(id)
);
