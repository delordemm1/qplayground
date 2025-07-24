-- +goose Up
-- # Add config_json column to automation_steps table

-- 1. Changes
--   - Add `config_json` column to `automation_steps` table
--   - Column type: jsonb with default value '{}'
--   - This will store step-level configuration like skip conditions

-- +goose StatementBegin
DO $$ 
BEGIN
    -- Add config_json column if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'automation_steps' 
        AND column_name = 'config_json'
    ) THEN
        ALTER TABLE automation_steps ADD COLUMN config_json jsonb DEFAULT '{}';
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE automation_steps DROP COLUMN IF EXISTS config_json;
-- +goose StatementEnd