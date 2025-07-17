-- +goose Up
/*
# Add personal organization reference to users table

1. Changes
  - Add `personal_organization_id` column to `users` table
  - Foreign key constraint to `organizations.id`
  - Nullable to allow for existing users without personal orgs initially
*/

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'personal_organization_id'
    ) THEN
        ALTER TABLE users ADD COLUMN personal_organization_id uuid REFERENCES organizations(id) ON DELETE SET NULL;
    END IF;
END $$;

-- +goose Down
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'personal_organization_id'
    ) THEN
        ALTER TABLE users DROP COLUMN personal_organization_id;
    END IF;
END $$;