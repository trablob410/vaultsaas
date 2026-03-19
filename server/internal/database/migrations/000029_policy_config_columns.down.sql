ALTER TABLE secrets  DROP COLUMN IF EXISTS policy_config;
ALTER TABLE projects DROP COLUMN IF EXISTS policy_config;
