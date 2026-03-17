-- Rename requester_id -> requester_user_id, add new columns, update check
ALTER TABLE access_requests
    RENAME COLUMN requester_id TO requester_user_id;

ALTER TABLE access_requests
    ADD COLUMN ai_agent_id               VARCHAR(255),
    ADD COLUMN rejection_reason           TEXT,
    ADD COLUMN requested_duration_minutes INTEGER NOT NULL DEFAULT 60;

-- Drop the old requester_type check and add updated one
ALTER TABLE access_requests
    DROP CONSTRAINT IF EXISTS access_requests_requester_type_check;
ALTER TABLE access_requests
    ADD CONSTRAINT access_requests_requester_type_check
        CHECK (requester_type IN ('human', 'ai_agent'));

-- Drop duration and updated_at
ALTER TABLE access_requests
    DROP COLUMN IF EXISTS duration,
    DROP COLUMN IF EXISTS updated_at;

-- Rename old index
DROP INDEX IF EXISTS idx_access_requests_requester_id;
CREATE INDEX idx_access_requests_requester_user_id ON access_requests (requester_user_id);
