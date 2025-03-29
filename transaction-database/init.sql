CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE stock_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stock_id UUID,
    parent_stock_transaction_id UUID,
    user_stock_transaction_id UUID,
    wallet_transaction_id UUID,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    order_status TEXT NOT NULL,
    is_buy BOOLEAN NOT NULL,
    order_type TEXT NOT NULL,
    stock_price DECIMAL NOT NULL,
    quantity INT NOT NULL,
    user_id UUID NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_stock_transaction_id) REFERENCES stock_transactions(id)
);

CREATE TABLE wallet_transactions (
    id uuid primary key default uuid_generate_v4(),
    stock_transaction_id uuid,
    wallet_id uuid,
    user_stock_transaction_id uuid,
    date_created timestamp default current_timestamp,
    date_modified timestamp default current_timestamp ON UPDATE CURRENT_TIMESTAMP,
    is_debit boolean not null,
    amount decimal not null,
    user_id uuid not null,
    timestamp timestamp default current_timestamp ON UPDATE CURRENT_TIMESTAMP,
    foreign key (stock_transaction_id) references stock_transactions(id)
);

-- -- Explicitly create the replication user if it doesn't exist
-- DO $$
-- BEGIN
--    -- Check if the role exists
--    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'trans_repl_user') THEN
--       -- Create the role using the password from the environment variable
--       -- Note: Using the actual password here is less secure than ENV VAR direct handling
--       -- But necessary if Bitnami's internal creation isn't working reliably.
--       -- Make sure 'trans_repl_password' below matches your .env value!
--       CREATE ROLE trans_repl_user WITH LOGIN REPLICATION PASSWORD 'trans_repl_password';
--    ELSE
--       -- Optionally, ensure it has the replication attribute if it somehow exists without it
--       ALTER ROLE trans_repl_user WITH REPLICATION;
--    END IF;
-- END $$;

-- -- Optionally grant connect if needed, usually not required for REPLICATION role itself
-- -- GRANT CONNECT ON DATABASE transaction_db TO trans_repl_user;