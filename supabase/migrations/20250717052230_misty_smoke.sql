-- +goose Up
/*
# Create automation steps table

1. New Tables
  - `automation_steps`
    - `id` (uuid, primary key, default gen_random_uuid())
    - `automation_id` (uuid, not null, foreign key to automations.id)
    - `name` (text, not null)
    - `step_order` (integer, not null) - for ordering steps
    - `created_at` (timestamptz, default now())
    - `updated_at` (timestamptz, default now())

2. Indexes
  - Index on automation_id and step_order for efficient ordering
*/

CREATE TABLE IF NOT EXISTS automation_steps (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    automation_id uuid NOT NULL,
    name text NOT NULL,
    step_order integer NOT NULL,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    FOREIGN KEY (automation_id) REFERENCES automations(id) ON DELETE CASCADE
);

-- Create index for efficient ordering
CREATE INDEX IF NOT EXISTS idx_automation_steps_automation_order 
    ON automation_steps(automation_id, step_order);

-- +goose Down
DROP INDEX IF EXISTS idx_automation_steps_automation_order;
DROP TABLE IF EXISTS automation_steps;