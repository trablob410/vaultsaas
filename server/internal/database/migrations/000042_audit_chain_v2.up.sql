-- Audit hash chain v2: strict total order + persisted chain head.
--
-- audit_chain_state holds the writer-side chain cursor. Log() serializes writers
-- with SELECT ... FOR UPDATE on this singleton row, assigns seq = last_seq + 1,
-- and updates last_hash in the same transaction. seq is therefore contiguous and
-- reflects true commit order, unlike (event_time, id).
--
-- The UNIQUE index is deliberately non-unique (plain index): audit_logs is
-- PARTITION BY RANGE (event_time) and a partitioned unique index must include
-- the partition key. Uniqueness of seq is enforced by the single-writer cursor
-- in audit_chain_state instead.

ALTER TABLE audit_logs ADD COLUMN seq BIGINT;

-- Metadata must round-trip byte-identical for the chain hash to verify after a
-- read-back. JSONB re-renders text (spaces, key order by length), so store the
-- exact bytes the hash was computed over. JSONB-specific queries are unused.
ALTER TABLE audit_logs ALTER COLUMN metadata TYPE TEXT USING metadata::text;

-- Backfill seq in existing insertion order (best effort: event_time, then id).
WITH ordered AS (
    SELECT id, event_time,
           ROW_NUMBER() OVER (ORDER BY event_time, id) AS rn
    FROM audit_logs
)
UPDATE audit_logs a
SET seq = o.rn
FROM ordered o
WHERE a.id = o.id AND a.event_time = o.event_time;

CREATE INDEX idx_audit_logs_seq ON audit_logs (seq);

CREATE TABLE audit_chain_state (
    id        INT  PRIMARY KEY,
    last_seq  BIGINT NOT NULL DEFAULT 0,
    last_hash TEXT   NOT NULL DEFAULT ''
);

-- Seed the cursor from the current tail of the chain (empty table -> seq 0, '').
INSERT INTO audit_chain_state (id, last_seq, last_hash)
SELECT 1,
       COALESCE((SELECT max(seq) FROM audit_logs), 0),
       COALESCE((SELECT hash_prev FROM audit_logs ORDER BY seq DESC NULLS LAST LIMIT 1), '')
ON CONFLICT (id) DO NOTHING;
