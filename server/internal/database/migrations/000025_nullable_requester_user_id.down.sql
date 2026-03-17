ALTER TABLE access_requests DROP CONSTRAINT IF EXISTS chk_requester_identity;
ALTER TABLE access_requests DROP CONSTRAINT IF EXISTS access_requests_requester_user_id_fkey;
ALTER TABLE access_requests
    ADD CONSTRAINT access_requests_requester_id_fkey
    FOREIGN KEY (requester_user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE access_requests ALTER COLUMN requester_user_id SET NOT NULL;
