DROP INDEX IF EXISTS idx_audit_logs_event_time;
DROP INDEX IF EXISTS idx_audit_logs_event_type;
DROP INDEX IF EXISTS idx_audit_logs_resource_type;
DROP INDEX IF EXISTS idx_audit_logs_user_id;

ALTER TABLE audit_logs
    DROP COLUMN IF EXISTS region_code,
    DROP COLUMN IF EXISTS user_agent,
    DROP COLUMN IF EXISTS ip_address,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS event_type;

ALTER TABLE audit_logs
    RENAME COLUMN resource_type TO resource;

ALTER TABLE audit_logs
    RENAME COLUMN user_id TO actor_id;

CREATE INDEX idx_audit_logs_actor ON audit_logs (actor_id);
CREATE INDEX idx_audit_logs_resource ON audit_logs (resource, resource_id);
