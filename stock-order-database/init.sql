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