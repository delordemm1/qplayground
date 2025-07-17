-- +goose Up
/*
# Create verification codes table for OTP authentication

1. New Tables
  - `verification_codes`
    - `id` (uuid, primary key, default gen_random_uuid())
    - `user_id` (uuid, nullable, foreign key to users.id)
    - `contact_method` (text, not null) - email or phone
    - `code` (text, not null) - the OTP code
    - `channel` (text, not null) - EMAIL, PHONE, WHATSAPP, TELEGRAM
    - `purpose` (text, not null) - AUTH
    - `expires_at` (timestamptz, not null)
    - `created_at` (timestamptz, default now())
*/

CREATE TABLE IF NOT EXISTS verification_codes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid,
    contact_method text NOT NULL,
    code text NOT NULL,
    channel text NOT NULL CHECK (channel IN ('EMAIL', 'PHONE', 'WHATSAPP', 'TELEGRAM')),
    purpose text NOT NULL CHECK (purpose IN ('AUTH')),
    expires_at timestamptz NOT NULL,
    created_at timestamptz DEFAULT now(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create index for code lookups
CREATE INDEX IF NOT EXISTS idx_verification_codes_contact_code ON verification_codes(contact_method, code, purpose);

-- Create index for cleanup of expired codes
CREATE INDEX IF NOT EXISTS idx_verification_codes_expires_at ON verification_codes(expires_at);

-- +goose Down
DROP INDEX IF EXISTS idx_verification_codes_contact_code;
DROP INDEX IF EXISTS idx_verification_codes_expires_at;
DROP TABLE IF EXISTS verification_codes;