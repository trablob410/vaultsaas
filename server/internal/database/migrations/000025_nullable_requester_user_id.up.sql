-- Make requester_user_id nullable (for agent-only access requests)
ALTER TABLE access_requests ALTER COLUMN requester_user_id DROP NOT NULL;

-- Drop old FK constraint (named after original column 'requester_id')
ALTER TABLE access_requests DROP CONSTRAINT IF EXISTS access_requests_requester_id_fkey;

-- Re-add FK with correct column name (nullable, agents have no user row)
ALTER TABLE access_requests
    ADD CONSTRAINT access_requests_requester_user_id_fkey
    FOREIGN KEY (requester_user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Ensure at least one requester identity is always set
ALTER TABLE access_requests
    ADD CONSTRAINT chk_requester_identity
    CHECK (requester_user_id IS NOT NULL OR ai_agent_id IS NOT NULL);
