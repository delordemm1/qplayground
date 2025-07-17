-- +goose Up
/*
# Create automation steps table

1. New Tables
  - `automation_steps`
    - `id` (uuid v7, primary key)
    - `automation_id` (uuid, not null, foreign key to automations.id)
    - `name` (text, not null)
    - `step_order` (integer, not null, for ordering steps)
    - `created_at` (timestamptz, default now)
    - `updated_at` (timestamptz, default now)

2. Security
  - Enable RLS on `automation_steps` table
  - Add policy for users to access steps in their automations

3. Indexes
  - Index on automation_id and step_order for efficient ordering
*/

CREATE TABLE IF NOT EXISTS automation_steps (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    automation_id uuid NOT NULL REFERENCES automations(id) ON DELETE CASCADE,
    name text NOT NULL,
    step_order integer NOT NULL,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

-- Create index for efficient ordering
CREATE INDEX IF NOT EXISTS idx_automation_steps_automation_order 
    ON automation_steps(automation_id, step_order);

ALTER TABLE automation_steps ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can manage steps in their automations"
    ON automation_steps
    FOR ALL
    TO authenticated
    USING (
        automation_id IN (
            SELECT a.id FROM automations a
            JOIN projects p ON a.project_id = p.id
            JOIN organizations o ON p.organization_id = o.id
            WHERE o.owner_user_id::text = auth.uid()::text
        )
    );

-- +goose Down
DROP POLICY IF EXISTS "Users can manage steps in their automations" ON automation_steps;
DROP INDEX IF EXISTS idx_automation_steps_automation_order;
DROP TABLE IF EXISTS automation_steps;