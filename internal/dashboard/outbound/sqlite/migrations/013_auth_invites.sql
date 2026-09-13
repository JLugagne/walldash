CREATE TABLE IF NOT EXISTS auth_invites (
    selector      TEXT PRIMARY KEY,
    verifier_hash TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'device',
    created_by    TEXT NOT NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at    TIMESTAMP NOT NULL,
    consumed_at   TIMESTAMP,
    consumed_by   TEXT,
    revoked_at    TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_auth_invites_exp ON auth_invites(expires_at);
