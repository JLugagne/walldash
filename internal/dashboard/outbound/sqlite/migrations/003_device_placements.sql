CREATE TABLE IF NOT EXISTS device_placements (
    id TEXT PRIMARY KEY,
    level_id TEXT NOT NULL REFERENCES levels(id) ON DELETE CASCADE,
    device_id TEXT NOT NULL,
    x REAL NOT NULL,
    y REAL NOT NULL,
    icon TEXT,
    custom_name TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_device_placements_level_id ON device_placements(level_id);
