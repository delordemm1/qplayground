-- +goose Up
-- # Add action_name column to automation_actions table

-- 1. Changes
--   - Add `action_name` column to `automation_actions` table
--   - Column type: text (nullable)
--   - This will store optional human-readable names for actions


-- +goose StatementBegin
DO $$ 
BEGIN
    -- Add action_name column if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'automation_actions' 
        AND column_name = 'action_name'
    ) THEN
        ALTER TABLE automation_actions ADD COLUMN action_name text;
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE automation_actions DROP COLUMN IF EXISTS action_name;
-- +goose StatementEnd