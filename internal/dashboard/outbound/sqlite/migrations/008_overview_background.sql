-- Per-dashboard background ambiance: an image URL (bundled sample or uploaded
-- file) rendered behind the Widget Grid, plus its display settings. Defaults keep
-- existing dashboards unchanged (no image) until one is selected.
ALTER TABLE overview_dashboards ADD COLUMN bg_image TEXT NOT NULL DEFAULT '';
ALTER TABLE overview_dashboards ADD COLUMN bg_opacity INTEGER NOT NULL DEFAULT 95;
ALTER TABLE overview_dashboards ADD COLUMN bg_blur INTEGER NOT NULL DEFAULT 14;
ALTER TABLE overview_dashboards ADD COLUMN bg_dim INTEGER NOT NULL DEFAULT 50;
