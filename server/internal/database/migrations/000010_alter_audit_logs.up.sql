-- Rename actor_id -> user_id, resource -> resource_type, add new columns
ALTER TABLE audit_logs
    RENAME COLUMN actor_id TO user_id;

ALTER TABLE audit_logs
    RENAME COLUMN resource TO resource_type;

ALTER TABLE audit_logs
    ADD COLUMN event_type   VARCHAR(50)  NOT NULL DEFAULT 'action',
    ADD COLUMN status       VARCHAR(20)  NOT NULL DEFAULT 'success',
    ADD COLUMN ip_address   INET,
    ADD COLUMN user_agent   TEXT,
    ADD COLUMN region_code  VARCHAR(2);

-- Recreate indexes with new column names
DROP INDEX IF EXISTS idx_audit_logs_actor;
DROP INDEX IF EXISTS idx_audit_logs_resource;
CREATE INDEX idx_audit_logs_user_id ON audit_logs (user_id);
CREATE INDEX idx_audit_logs_resource_type ON audit_logs (resource_type, resource_id);
CREATE INDEX idx_audit_logs_event_type ON audit_logs (event_type);
CREATE INDEX idx_audit_logs_event_time ON audit_logs (event_time);
