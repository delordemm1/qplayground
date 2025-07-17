-- +goose Up
/*
# Create user active sessions table

1. New Tables
  - `user_active_sessions`
    - `id` (uuid, primary key, default gen_random_uuid())
    - `user_id` (uuid, not null, foreign key to users.id)
    - `session_token` (text, not null)
    - `user_agent` (text, not null)
    - `ip_address` (text, not null)
    - `last_active_at` (timestamptz, default now())
    - `created_at` (timestamptz, default now())
*/

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_active_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    session_token text NOT NULL,
    user_agent text NOT NULL,
    ip_address text NOT NULL,
    last_active_at timestamptz DEFAULT now(),
    created_at timestamptz DEFAULT now(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create index for user lookups
CREATE INDEX IF NOT EXISTS idx_user_active_sessions_user_id ON user_active_sessions(user_id);

-- Create index for session token lookups
CREATE INDEX IF NOT EXISTS idx_user_active_sessions_token ON user_active_sessions(session_token);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_user_active_sessions_user_id;
DROP INDEX IF EXISTS idx_user_active_sessions_token;
DROP TABLE IF EXISTS user_active_sessions;
-- +goose StatementEnd
