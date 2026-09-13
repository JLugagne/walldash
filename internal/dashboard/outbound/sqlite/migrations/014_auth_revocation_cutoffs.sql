-- Account-scoped access-token revocation cutoffs.
--
-- egauth's access-token revocation tracker is an in-process map, so it is rebuilt empty on every
-- boot and an already-issued access JWT would be accepted again after a restart for the remainder
-- of its TTL. Walldash persists the cutoff here and replays it into the tracker at startup.
--
-- cutoff_at is stored as Unix nanoseconds rather than a TIMESTAMP so ordering and pruning never
-- depend on the driver's time formatting or on the stored timezone.
CREATE TABLE IF NOT EXISTS auth_revocation_cutoffs (
    user_id   TEXT PRIMARY KEY,
    cutoff_at INTEGER NOT NULL,
    reason    TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_auth_revocation_cutoffs_cutoff_at
    ON auth_revocation_cutoffs (cutoff_at);
