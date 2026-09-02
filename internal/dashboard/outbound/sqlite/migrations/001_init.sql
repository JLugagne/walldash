CREATE TABLE IF NOT EXISTS health_check (
    id INTEGER PRIMARY KEY,
    status TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

INSERT OR IGNORE INTO health_check (id, status, updated_at) VALUES (1, 'ok', CURRENT_TIMESTAMP);
