-- +goose Up
/*
# Create automations table

1. New Tables
  - `automations`
    - `id` (uuid v7, primary key)
    - `project_id` (uuid, not null, foreign key to projects.id)
    - `name` (text, not null)
    - `description` (text, nullable)
    - `config_json` (jsonb, stores variables, run settings, templates, etc.)
    - `created_at` (timestamptz, default now)
    - `updated_at` (timestamptz, default now)

2. Security
  - Enable RLS on `automations` table
  - Add policy for users to access automations in their projects
*/

CREATE TABLE IF NOT EXISTS automations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name text NOT NULL,
    description text,
    config_json jsonb DEFAULT '{}',
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

ALTER TABLE automations ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can manage automations in their projects"
    ON automations
    FOR ALL
    TO authenticated
    USING (
        project_id IN (
            SELECT p.id FROM projects p
            JOIN organizations o ON p.organization_id = o.id
            WHERE o.owner_user_id::text = auth.uid()::text
        )
    );

-- +goose Down
DROP POLICY IF EXISTS "Users can manage automations in their projects" ON automations;
DROP TABLE IF EXISTS automations;