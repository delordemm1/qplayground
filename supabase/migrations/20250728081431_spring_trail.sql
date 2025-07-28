-- +goose Up
/*
# Add report_urls columns to automation_runs table

1. Changes
  - Add `detailed_report_url` column to `automation_runs` table
  - Add `user_journey_report_url` column to `automation_runs` table
  - Column type: text (nullable)
  - These will store the public URLs of the generated HTML reports.
*/

-- +goose StatementBegin
DO $$ 
BEGIN
    -- Add detailed_report_url column if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'automation_runs' 
        AND column_name = 'detailed_report_url'
    ) THEN
        ALTER TABLE automation_runs ADD COLUMN detailed_report_url text;
    END IF;

    -- Add user_journey_report_url column if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'automation_runs' 
        AND column_name = 'user_journey_report_url'
    ) THEN
        ALTER TABLE automation_runs ADD COLUMN user_journey_report_url text;
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE automation_runs DROP COLUMN IF EXISTS detailed_report_url;
ALTER TABLE automation_runs DROP COLUMN IF EXISTS user_journey_report_url;
-- +goose StatementEnd