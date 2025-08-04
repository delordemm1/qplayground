-- +goose Up
/*
# Add multi-runner support to automation_runs table

1. Changes
  - Add `sub_run_outputs_json` column to `automation_runs` table
  - Add `total_runs_expected` column to `automation_runs` table  
  - Add `runs_completed` column to `automation_runs` table
  - Update status check constraint to include new statuses
*/

-- +goose StatementBegin
DO $$ 
BEGIN
    -- Add sub_run_outputs_json column if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'automation_runs' 
        AND column_name = 'sub_run_outputs_json'
    ) THEN
        ALTER TABLE automation_runs ADD COLUMN sub_run_outputs_json jsonb DEFAULT '{}';
    END IF;

    -- Add total_runs_expected column if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'automation_runs' 
        AND column_name = 'total_runs_expected'
    ) THEN
        ALTER TABLE automation_runs ADD COLUMN total_runs_expected integer;
    END IF;

    -- Add runs_completed column if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'automation_runs' 
        AND column_name = 'runs_completed'
    ) THEN
        ALTER TABLE automation_runs ADD COLUMN runs_completed integer DEFAULT 0;
    END IF;

    -- Update status constraint to include new statuses
    ALTER TABLE automation_runs DROP CONSTRAINT IF EXISTS automation_runs_status_check;
    ALTER TABLE automation_runs ADD CONSTRAINT automation_runs_status_check 
        CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled', 'queued', 'partial_completed', 'awaiting_external_runner', 'consolidating'));
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE automation_runs DROP COLUMN IF EXISTS sub_run_outputs_json;
ALTER TABLE automation_runs DROP COLUMN IF EXISTS total_runs_expected;
ALTER TABLE automation_runs DROP COLUMN IF EXISTS runs_completed;

-- Restore original status constraint
ALTER TABLE automation_runs DROP CONSTRAINT IF EXISTS automation_runs_status_check;
ALTER TABLE automation_runs ADD CONSTRAINT automation_runs_status_check 
    CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled', 'queued'));
-- +goose StatementEnd