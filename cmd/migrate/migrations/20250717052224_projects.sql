-- +goose Up
/*
# Create projects table

1. New Tables
  - `projects`
    - `id` (uuid, primary key, default gen_random_uuid())
    - `organization_id` (uuid, not null, foreign key to organizations.id)
    - `name` (text, not null)
    - `description` (text, nullable)
    - `created_at` (timestamptz, default now())
    - `updated_at` (timestamptz, default now())
*/

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS projects (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL,
    name text NOT NULL,
    description text,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

-- Create index for organization lookups
CREATE INDEX IF NOT EXISTS idx_projects_organization_id ON projects(organization_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_projects_organization_id;
DROP TABLE IF EXISTS projects;
-- +goose StatementEnd
