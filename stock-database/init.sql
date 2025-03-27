CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE stocks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    Name TEXT NOT NULL
);

-- CREATE OR REPLACE FUNCTION update_date_modified()
-- RETURNS TRIGGER AS $$
-- BEGIN
--     -- On an UPDATE, set the new date_modified to the current time
--     NEW.date_modified := CURRENT_TIMESTAMP;
--     RETURN NEW;
-- END;
-- $$ LANGUAGE plpgsql;
-- CREATE TRIGGER set_date_modified
-- BEFORE UPDATE ON stocks
-- FOR EACH ROW
-- EXECUTE PROCEDURE update_date_modified();
