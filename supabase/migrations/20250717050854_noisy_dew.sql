-- +goose Up
/*
# Create projects table

1. New Tables
  - `projects`
    - `id` (uuid v7, primary key)
    - `organization_id` (uuid, not null, foreign key to organizations.id)
    - `name` (text, not null)
    - `description` (text, nullable)
    - `created_at` (timestamptz, default now)
    - `updated_at` (timestamptz, default now)

2. Security
  - Enable RLS on `projects` table
  - Add policy for users to access projects in their organizations
*/

CREATE TABLE IF NOT EXISTS projects (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name text NOT NULL,
    description text,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

ALTER TABLE projects ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can manage projects in their organizations"
    ON projects
    FOR ALL
    TO authenticated
    USING (
        organization_id IN (
            SELECT id FROM organizations WHERE owner_user_id::text = auth.uid()::text
        )
    );

-- +goose Down
DROP POLICY IF EXISTS "Users can manage projects in their organizations" ON projects;
DROP TABLE IF EXISTS projects;