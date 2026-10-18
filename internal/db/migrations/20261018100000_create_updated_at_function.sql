-- +goose Up
-- Shared trigger function: keeps updated_at correct even if a query forgets to set it.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
DROP FUNCTION IF EXISTS set_updated_at();
