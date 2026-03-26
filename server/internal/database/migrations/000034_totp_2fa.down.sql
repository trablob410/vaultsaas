DROP TABLE IF EXISTS backup_codes;

ALTER TABLE users
  DROP COLUMN IF EXISTS totp_secret_enc,
  DROP COLUMN IF EXISTS totp_enabled,
  DROP COLUMN IF EXISTS totp_setup_secret;
