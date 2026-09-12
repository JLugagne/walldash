CREATE TABLE IF NOT EXISTS auth_accounts (
    id           TEXT PRIMARY KEY,
    status       TEXT NOT NULL DEFAULT 'active',
    role         TEXT NOT NULL DEFAULT 'device',
    label        TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen_at TIMESTAMP,
    revoked_at   TIMESTAMP
);

CREATE TABLE IF NOT EXISTS auth_refresh_tokens (
    hash                 TEXT PRIMARY KEY,
    family_id            TEXT NOT NULL,
    user_id              TEXT NOT NULL,
    tenant_id            TEXT NOT NULL DEFAULT '',
    auth_time            TIMESTAMP NOT NULL,
    must_change_password INTEGER NOT NULL DEFAULT 0,
    expires_at           TIMESTAMP NOT NULL,
    created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    consumed_at          TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_auth_rt_user   ON auth_refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_auth_rt_family ON auth_refresh_tokens(family_id);
CREATE INDEX IF NOT EXISTS idx_auth_rt_exp    ON auth_refresh_tokens(expires_at);

CREATE TABLE IF NOT EXISTS auth_api_keys (
    id         TEXT PRIMARY KEY,
    tenant_id  TEXT NOT NULL DEFAULT '',
    hash       TEXT NOT NULL UNIQUE,
    prefix     TEXT NOT NULL DEFAULT '',
    type       TEXT NOT NULL DEFAULT '',
    created_by TEXT NOT NULL,
    claims     BLOB NOT NULL,
    expires_at TIMESTAMP,
    revoked_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_auth_api_keys_created_by ON auth_api_keys(created_by);

CREATE TABLE IF NOT EXISTS app_secrets (
    name       TEXT PRIMARY KEY,
    value      BLOB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
