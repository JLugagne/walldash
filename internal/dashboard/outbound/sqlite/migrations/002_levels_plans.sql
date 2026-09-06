CREATE TABLE IF NOT EXISTS levels (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    "order" INTEGER NOT NULL,
    is_outdoor BOOLEAN NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS plans (
    level_id TEXT PRIMARY KEY,
    walls_json TEXT NOT NULL DEFAULT '[]',
    zones_json TEXT NOT NULL DEFAULT '[]',
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY(level_id) REFERENCES levels(id) ON DELETE CASCADE
);
