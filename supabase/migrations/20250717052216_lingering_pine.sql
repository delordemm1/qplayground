-- +goose Up
/*
# Create organizations table

1. New Tables
  - `organizations`
    - `id` (uuid, primary key, default gen_random_uuid())
    - `name` (text, not null)
    - `owner_user_id` (uuid, not null, foreign key to users.id)
    - `created_at` (timestamptz, default now())
    - `updated_at` (timestamptz, default now())
*/

CREATE TABLE IF NOT EXISTS organizations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    owner_user_id uuid NOT NULL,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create index for owner lookups
CREATE INDEX IF NOT EXISTS idx_organizations_owner_user_id ON organizations(owner_user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_organizations_owner_user_id;
DROP TABLE IF EXISTS organizations;