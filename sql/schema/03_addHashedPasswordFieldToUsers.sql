-- +goose Up
ALTER TABLE users ADD COLUMN hashed_password VARCHAR(255) NOT NULL DEFAULT 'unset';

-- +goose Down
ALTER TABLEW users DROP COLUMN hashed_password;