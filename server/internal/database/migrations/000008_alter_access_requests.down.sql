DROP INDEX IF EXISTS idx_access_requests_requester_user_id;

ALTER TABLE access_requests
    ADD COLUMN IF NOT EXISTS duration   INTERVAL NOT NULL DEFAULT '1 hour',
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE access_requests
    DROP COLUMN IF EXISTS requested_duration_minutes,
    DROP COLUMN IF EXISTS rejection_reason,
    DROP COLUMN IF EXISTS ai_agent_id;

ALTER TABLE access_requests
    DROP CONSTRAINT IF EXISTS access_requests_requester_type_check;
ALTER TABLE access_requests
    ADD CONSTRAINT access_requests_requester_type_check
        CHECK (requester_type IN ('human', 'agent'));

ALTER TABLE access_requests
    RENAME COLUMN requester_user_id TO requester_id;

CREATE INDEX idx_access_requests_requester_id ON access_requests (requester_id);
