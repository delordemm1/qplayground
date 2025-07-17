-- +goose Up
/*
# Create organizations table

1. New Tables
  - `organizations`
    - `id` (uuid v7, primary key)
    - `name` (text, not null)
    - `owner_user_id` (uuid, not null, foreign key to users.id)
    - `created_at` (timestamptz, default now)
    - `updated_at` (timestamptz, default now)

2. Security
  - Enable RLS on `organizations` table
  - Add policy for users to read/write their own organizations
*/

CREATE TABLE IF NOT EXISTS organizations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    owner_user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can manage their own organizations"
    ON organizations
    FOR ALL
    TO authenticated
    USING (auth.uid()::text = owner_user_id::text);

-- +goose Down
DROP POLICY IF EXISTS "Users can manage their own organizations" ON organizations;
DROP TABLE IF EXISTS organizations;