-- +goose Up
/*
# Create sessions table for SCS session management

1. New Tables
  - `sessions`
    - `token` (text, primary key)
    - `data` (bytea, not null)
    - `expiry` (timestamptz, not null)
*/

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS sessions (
    token text PRIMARY KEY,
    data bytea NOT NULL,
    expiry timestamptz NOT NULL
);
-- +goose StatementEnd

-- Create index for expiry cleanup
CREATE INDEX IF NOT EXISTS idx_sessions_expiry ON sessions(expiry);

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_sessions_expiry;
DROP TABLE IF EXISTS sessions;
-- +goose StatementEnd
