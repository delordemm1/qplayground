-- +goose Up
/*
# Create automations table

1. New Tables
  - `automations`
    - `id` (uuid, primary key, default gen_random_uuid())
    - `project_id` (uuid, not null, foreign key to projects.id)
    - `name` (text, not null)
    - `description` (text, nullable)
    - `config_json` (jsonb, default '{}') - stores variables, run settings, templates
    - `created_at` (timestamptz, default now())
    - `updated_at` (timestamptz, default now())
*/

CREATE TABLE IF NOT EXISTS automations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL,
    name text NOT NULL,
    description text,
    config_json jsonb DEFAULT '{}',
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Create index for project lookups
CREATE INDEX IF NOT EXISTS idx_automations_project_id ON automations(project_id);

-- +goose Down
DROP INDEX IF EXISTS idx_automations_project_id;
DROP TABLE IF EXISTS automations;