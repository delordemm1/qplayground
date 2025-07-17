-- +goose Up
/*
# Create automation actions table

1. New Tables
  - `automation_actions`
    - `id` (uuid v7, primary key)
    - `step_id` (uuid, not null, foreign key to automation_steps.id)
    - `action_type` (text, e.g., "playwright:goto", "playwright:click")
    - `action_config_json` (jsonb, stores specific parameters for the action)
    - `action_order` (integer, not null, for ordering actions within a step)
    - `created_at` (timestamptz, default now)
    - `updated_at` (timestamptz, default now)

2. Security
  - Enable RLS on `automation_actions` table
  - Add policy for users to access actions in their steps

3. Indexes
  - Index on step_id and action_order for efficient ordering
  - Index on action_type for filtering by plugin type
*/

CREATE TABLE IF NOT EXISTS automation_actions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    step_id uuid NOT NULL REFERENCES automation_steps(id) ON DELETE CASCADE,
    action_type text NOT NULL,
    action_config_json jsonb DEFAULT '{}',
    action_order integer NOT NULL,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_automation_actions_step_order 
    ON automation_actions(step_id, action_order);

CREATE INDEX IF NOT EXISTS idx_automation_actions_type 
    ON automation_actions(action_type);

ALTER TABLE automation_actions ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can manage actions in their steps"
    ON automation_actions
    FOR ALL
    TO authenticated
    USING (
        step_id IN (
            SELECT s.id FROM automation_steps s
            JOIN automations a ON s.automation_id = a.id
            JOIN projects p ON a.project_id = p.id
            JOIN organizations o ON p.organization_id = o.id
            WHERE o.owner_user_id::text = auth.uid()::text
        )
    );

-- +goose Down
DROP POLICY IF EXISTS "Users can manage actions in their steps" ON automation_actions;
DROP INDEX IF EXISTS idx_automation_actions_step_order;
DROP INDEX IF EXISTS idx_automation_actions_type;
DROP TABLE IF EXISTS automation_actions;