CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE stockOrder (
    id uuid primary key default uuid_generate_v4(),
    stock_id uuid,
    parent_stock_order_id uuid,
    date_created timestamp default current_timestamp,
    date_modified timestamp default current_timestamp,
    is_buy boolean not null,
    order_type text not null,
    price decimal not null,
    quantity int not null,
    user_id uuid not null,
    foreign key (parent_stock_order_id) references stockorder(id)
);

-- Create a function that deletes the row if quantity is zero
CREATE OR REPLACE FUNCTION delete_when_quantity_zero()
RETURNS TRIGGER AS $$
BEGIN
    -- If the updated quantity is 0, remove the row
    IF NEW.quantity = 0 THEN
        DELETE FROM stockOrder WHERE id = OLD.id;
        -- Return NULL to indicate no further update for this row
        RETURN NULL;
    END IF;

    -- Otherwise, proceed with the update
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create a trigger that calls the above function before any update
CREATE TRIGGER quantity_zero_deletion_trigger
BEFORE UPDATE
ON stockOrder
FOR EACH ROW
EXECUTE PROCEDURE delete_when_quantity_zero();