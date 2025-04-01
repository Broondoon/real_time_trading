CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable Citus extension
CREATE EXTENSION IF NOT EXISTS citus;

-- Create the Users table as a distributed table
CREATE TABLE Users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) NOT NULL UNIQUE,
    password TEXT NOT NULL,
    name TEXT NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- Distribute the table by id (hash distribution)
SELECT create_distributed_table('Users', 'id');

