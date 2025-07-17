-- +goose Up
/*
# Create automation runs table

1. New Tables
  - `automation_runs`
    - `id` (uuid v7, primary key)
    - `automation_id` (uuid, not null, foreign key to automations.id)
    - `status` (text, e.g., "pending", "running", "completed", "failed", "cancelled")
    - `start_time` (timestamptz, nullable)
    - `end_time` (timestamptz, nullable)
    - `logs_json` (jsonb, stores execution logs)
    - `output_files_json` (jsonb, stores paths/URLs to screenshots, etc.)
    - `error_message` (text, nullable, stores error details if failed)
    - `created_at` (timestamptz, default now)
    - `updated_at` (timestamptz, default now)

2. Security
  - Enable RLS on `automation_runs` table
  - Add policy for users to access runs of their automations

3. Indexes
  - Index on automation_id for efficient querying
  - Index on status for filtering by run status
  - Index on created_at for chronological ordering
*/

CREATE TABLE IF NOT EXISTS automation_runs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    automation_id uuid NOT NULL REFERENCES automations(id) ON DELETE CASCADE,
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled')),
    start_time timestamptz,
    end_time timestamptz,
    logs_json jsonb DEFAULT '[]',
    output_files_json jsonb DEFAULT '[]',
    error_message text,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_automation_runs_automation_id 
    ON automation_runs(automation_id);

CREATE INDEX IF NOT EXISTS idx_automation_runs_status 
    ON automation_runs(status);

CREATE INDEX IF NOT EXISTS idx_automation_runs_created_at 
    ON automation_runs(created_at DESC);

ALTER TABLE automation_runs ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view runs of their automations"
    ON automation_runs
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
DROP POLICY IF EXISTS "Users can view runs of their automations" ON automation_runs;
DROP INDEX IF EXISTS idx_automation_runs_automation_id;
DROP INDEX IF EXISTS idx_automation_runs_status;
DROP INDEX IF EXISTS idx_automation_runs_created_at;
DROP TABLE IF EXISTS automation_runs;