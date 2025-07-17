-- +goose Up
/*
# Create users table

1. New Tables
  - `users`
    - `id` (uuid, primary key, default gen_random_uuid())
    - `email` (text, unique, not null)
    - `role` (text, default 'USER')
    - `sub` (text, not null) - for OAuth/auth providers
    - `avatar` (text, nullable)
    - `current_org_id` (uuid, nullable) - for dynamic role switching
    - `personal_organization_id` (uuid, nullable) - link to personal org
    - `deleted_at` (timestamptz, nullable)
    - `created_at` (timestamptz, default now())
    - `updated_at` (timestamptz, default now())
*/

CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text UNIQUE NOT NULL,
    role text DEFAULT 'USER' CHECK (role IN ('USER', 'ADMIN')),
    sub text NOT NULL,
    avatar text,
    current_org_id uuid,
    personal_organization_id uuid,
    deleted_at timestamptz,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

-- Create index for email lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Create index for sub lookups (OAuth)
CREATE INDEX IF NOT EXISTS idx_users_sub ON users(sub);

-- +goose Down
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_sub;
DROP TABLE IF EXISTS users;