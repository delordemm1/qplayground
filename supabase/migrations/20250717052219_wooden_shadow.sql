-- +goose Up
/*
# Add foreign key constraints to users table

1. Changes
  - Add foreign key constraint for current_org_id
  - Add foreign key constraint for personal_organization_id
*/

-- Add foreign key constraints to users table
DO $$
BEGIN
    -- Add foreign key for current_org_id if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_users_current_org_id' 
        AND table_name = 'users'
    ) THEN
        ALTER TABLE users ADD CONSTRAINT fk_users_current_org_id 
            FOREIGN KEY (current_org_id) REFERENCES organizations(id) ON DELETE SET NULL;
    END IF;

    -- Add foreign key for personal_organization_id if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_users_personal_organization_id' 
        AND table_name = 'users'
    ) THEN
        ALTER TABLE users ADD CONSTRAINT fk_users_personal_organization_id 
            FOREIGN KEY (personal_organization_id) REFERENCES organizations(id) ON DELETE SET NULL;
    END IF;
END $$;

-- +goose Down
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_current_org_id;
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_personal_organization_id;