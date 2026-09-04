DROP TABLE IF EXISTS audit_chain_state;
DROP INDEX IF EXISTS idx_audit_logs_seq;
ALTER TABLE audit_logs ALTER COLUMN metadata TYPE JSONB USING metadata::jsonb;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS seq;
