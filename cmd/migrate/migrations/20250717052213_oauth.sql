-- +goose Up
/*
# Create OAuth states table for OAuth authentication

1. New Tables
  - `oauth_states`
    - `state` (text, primary key)
    - `provider` (text, not null) - google, facebook, github, etc.
    - `user_id` (uuid, nullable, foreign key to users.id)
    - `verifier` (text, not null)
    - `expires_at` (timestamptz, not null)
    - `created_at` (timestamptz, default now())
    - `updated_at` (timestamptz, default now())
*/

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS oauth_states (
    state text PRIMARY KEY,
    provider text NOT NULL CHECK (provider IN ('google', 'facebook', 'github', 'x', 'linkedin')),
    user_id uuid,
    verifier text NOT NULL,
    expires_at timestamptz NOT NULL,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create index for cleanup of expired states
CREATE INDEX IF NOT EXISTS idx_oauth_states_expires_at ON oauth_states(expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_oauth_states_expires_at;
DROP TABLE IF EXISTS oauth_states;
-- +goose StatementEnd
