-- First-run starter: on a brand-new, still empty database, seed one dashboard with a
-- 5-day weather forecast so the user lands on something useful instead of the "No
-- Dashboard" empty state. The weather widget resolves its location and units from the
-- Home Assistant instance configuration (no override), and its caption falls back to the
-- instance location name.
--
-- Guarded on an empty dashboards table, so upgrading an existing install never adds a
-- dashboard the user did not ask for. Because a migration is applied at most once, a user
-- who later deletes every dashboard does not get this one back.
INSERT INTO dashboards (id, name, "order", cols, rows, bg_image, bg_opacity, bg_blur, bg_dim, created_at, updated_at)
SELECT '00000000-0000-0000-0000-000000000001', 'Home', 0, 12, 8, '/backgrounds/desert-night.jpg', 95, 14, 50, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM dashboards);

INSERT INTO widgets (id, dashboard_id, type, title, "order", config_json, col, row, col_span, row_span, created_at, updated_at)
SELECT '00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'weather', '', 0,
       json_object('display', 'weather', 'weather_mode', 'ndays', 'weather_days', 5),
       0, 0, 2, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
WHERE EXISTS (SELECT 1 FROM dashboards WHERE id = '00000000-0000-0000-0000-000000000001')
  AND NOT EXISTS (SELECT 1 FROM widgets);
