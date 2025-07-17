-- +goose Up
/*
# Create automation actions table

1. New Tables
  - `automation_actions`
    - `id` (uuid, primary key, default gen_random_uuid())
    - `step_id` (uuid, not null, foreign key to automation_steps.id)
    - `action_type` (text, not null) - e.g., "playwright:goto", "playwright:click"
    - `action_config_json` (jsonb, default '{}') - stores action-specific parameters
    - `action_order` (integer, not null) - for ordering actions within a step
    - `created_at` (timestamptz, default now())
    - `updated_at` (timestamptz, default now())

2. Indexes
  - Index on step_id and action_order for efficient ordering
  - Index on action_type for filtering by plugin type
*/

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS automation_actions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    step_id uuid NOT NULL,
    action_type text NOT NULL,
    action_config_json jsonb DEFAULT '{}',
    action_order integer NOT NULL,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    FOREIGN KEY (step_id) REFERENCES automation_steps(id) ON DELETE CASCADE
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_automation_actions_step_order 
    ON automation_actions(step_id, action_order);

CREATE INDEX IF NOT EXISTS idx_automation_actions_type 
    ON automation_actions(action_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_automation_actions_step_order;
DROP INDEX IF EXISTS idx_automation_actions_type;
DROP TABLE IF EXISTS automation_actions;
-- +goose StatementEnd
